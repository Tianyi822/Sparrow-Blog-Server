package webservice

import (
	"context"
	"errors"
	"fmt"
	"sparrow_blog_server/cache"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/vo"
	"sparrow_blog_server/internal/repositories/blogrepo"
	"sparrow_blog_server/internal/repositories/categoryrepo"
	"sparrow_blog_server/internal/repositories/commentrepo"
	"sparrow_blog_server/internal/repositories/friendlinkrepo"
	"sparrow_blog_server/internal/repositories/tagrepo"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"sparrow_blog_server/storage/ossstore"
	"time"
)

// GetHomeData 获取首页数据。
// 该函数通过查询配置和数据库来收集用户信息、博客、分类和标签等数据。
// 参数:
//   - ctx context.Context: 上下文对象，用于传递请求范围的 deadline、取消信号等。
//
// 返回值:
//   - map[string]any: 包含首页所需数据的映射，包括用户信息、博客、分类和标签等。
//   - error: 如果查询数据过程中发生错误，则返回该错误。
func GetHomeData(ctx context.Context) (map[string]any, error) {
	// 初始化结果映射，填充用户配置信息。
	result := map[string]any{
		"user_name":           config.User.Username,
		"user_email":          config.User.UserEmail,
		"user_github_address": config.User.UserGithubAddress,
		"user_hobbies":        config.User.UserHobbies,
		"type_writer_content": config.User.TypeWriterContent,
		"background_image":    config.User.BackgroundImage,
		"avatar_image":        config.User.AvatarImage,
		"web_logo":            config.User.WebLogo,
		"icp_filing_number":   config.User.ICPFilingNumber,
	}

	// 定义一个结构体用于存储查询结果和可能的错误。
	type resultData struct {
		Blogs      any
		Categories any
		Tags       any
		Err        error
	}

	// 创建一个带缓冲的通道，用于接收查询结果。
	ch := make(chan resultData, 3)

	// 启动三个协程，分别查询博客、分类和标签数据。
	go func() {
		blogDtos, err := blogrepo.FindAllBlogs(ctx, true)
		if err != nil {
			ch <- resultData{Err: fmt.Errorf("failed to find all blogs: %w", err)}
			return
		}

		var blogVos []dto.BlogDto
		for _, blogDto := range blogDtos {
			cat, err := categoryrepo.FindCategoryById(ctx, blogDto.CategoryId)
			if err != nil {
				ch <- resultData{Err: fmt.Errorf("failed to find category by id: %w", err)}
				return
			}

			tags, err := tagrepo.FindTagsByBlogId(ctx, blogDto.BlogId)
			if err != nil {
				ch <- resultData{Err: fmt.Errorf("failed to find tags by blog id: %w", err)}
				return
			}

			blogVo := dto.BlogDto{
				BlogId:       blogDto.BlogId,
				BlogTitle:    blogDto.BlogTitle,
				BlogImageId:  blogDto.BlogImageId,
				BlogBrief:    blogDto.BlogBrief,
				BlogWordsNum: blogDto.BlogWordsNum,
				BlogIsTop:    blogDto.BlogIsTop,
				BlogState:    blogDto.BlogState,
				Category:     cat,
				Tags:         tags,
				CreateTime:   blogDto.CreateTime,
				UpdateTime:   blogDto.UpdateTime,
			}
			blogVos = append(blogVos, blogVo)
		}

		ch <- resultData{Blogs: blogVos}
	}()

	go func() {
		categories, err := categoryrepo.FindAllCategories(ctx)
		if err != nil {
			ch <- resultData{Err: fmt.Errorf("failed to get all categories: %w", err)}
			return
		}

		var cateVos []dto.CategoryDto
		for _, cate := range categories {
			cateVo := dto.CategoryDto{
				CategoryId:   cate.CategoryId,
				CategoryName: cate.CategoryName,
			}
			cateVos = append(cateVos, cateVo)
		}

		ch <- resultData{Categories: cateVos}
	}()

	go func() {
		tags, err := tagrepo.FindAllTags(ctx)
		if err != nil {
			ch <- resultData{Err: fmt.Errorf("failed to get all tags: %w", err)}
			return
		}

		cateVos := make([]dto.TagDto, 0, len(tags))
		for _, tag := range tags {
			cateVo := dto.TagDto{
				TagId:   tag.TagId,
				TagName: tag.TagName,
			}
			cateVos = append(cateVos, cateVo)
		}

		ch <- resultData{Tags: cateVos}
	}()

	for i := 0; i < 3; i++ {
		r := <-ch
		if r.Err != nil {
			return nil, r.Err
		}
		if r.Blogs != nil {
			result["blogs"] = r.Blogs
		} else if r.Categories != nil {
			result["categories"] = r.Categories
		} else if r.Tags != nil {
			result["tags"] = r.Tags
		}
	}

	// 返回包含所有首页数据的映射。
	return result, nil
}

// GetBlogDataById 根据博客ID获取博客详细数据。
// 该函数通过博客ID查询博客信息，包括博客基本信息、分类信息、标签信息以及博客内容的预签名URL。
//
// 参数:
//   - ctx context.Context: 上下文对象，用于传递请求范围的 deadline、取消信号等
//   - id string: 博客的唯一标识符
//
// 返回值:
//   - *vo.BlogVo: 包含博客详细信息的视图对象，包括博客基本信息、分类和标签
//   - string: 博客内容的预签名URL，用于访问存储在对象存储中的博客内容
//   - error: 如果查询过程中发生错误，则返回该错误
//
// 函数逻辑:
// 1. 根据博客ID查询博客基本信息
// 2. 如果博客存在:
//   - 查询博客关联的分类信息
//   - 查询博客关联的标签信息
//   - 构建博客视图对象(BlogVo)
//   - 尝试从缓存获取预签名URL
//   - 如果缓存未命中:
//   - 生成OSS存储路径
//   - 生成新的预签名URL(有效期20分钟)
//   - 将URL存入缓存
//
// 3. 如果博客不存在:
//   - 记录警告日志
//   - 返回错误信息
func GetBlogDataById(ctx context.Context, id string) (*vo.BlogVo, string, error) {
	// 根据ID查询博客信息
	blogDto, err := blogrepo.FindBlogById(ctx, id)
	if err != nil {
		return nil, "", err
	}

	var blogVo *vo.BlogVo
	var preUrl string

	if blogDto != nil {
		// 查询博客关联的分类信息
		cat, err := categoryrepo.FindCategoryById(ctx, blogDto.CategoryId)
		if err != nil {
			return nil, "", err
		}
		// 构建分类视图对象
		catVo := &vo.CategoryVo{
			CategoryId:   cat.CategoryId,
			CategoryName: cat.CategoryName,
		}

		// 查询博客关联的标签信息
		tags, err := tagrepo.FindTagsByBlogId(ctx, blogDto.BlogId)
		if err != nil {
			return nil, "", err
		}
		// 构建标签视图对象列表
		var tagVos []vo.TagVo
		for _, tag := range tags {
			tagVo := vo.TagVo{
				TagId:   tag.TagId,
				TagName: tag.TagName,
			}
			tagVos = append(tagVos, tagVo)
		}

		// 构建博客视图对象，包含基本信息、分类和标签
		blogVo = &vo.BlogVo{
			BlogId:       blogDto.BlogId,
			BlogTitle:    blogDto.BlogTitle,
			BlogImageId:  blogDto.BlogImageId,
			BlogBrief:    blogDto.BlogBrief,
			BlogWordsNum: blogDto.BlogWordsNum,
			BlogIsTop:    blogDto.BlogIsTop,
			BlogState:    blogDto.BlogState,
			Category:     catVo,
			Tags:         tagVos,
			CreateTime:   blogDto.CreateTime,
			UpdateTime:   blogDto.UpdateTime,
		}

		// 尝试从缓存获取预签名URL
		preUrl, err = storage.Storage.Cache.GetString(ctx, storage.BuildBlogCacheKey(blogDto.BlogId))
		if errors.Is(err, cache.ErrNotFound) {
			// 缓存未命中，生成OSS存储路径
			ossPath := ossstore.GenOssSavePath(blogDto.BlogTitle, ossstore.MarkDown)

			// 生成新的预签名URL，有效期20分钟
			presign, err := storage.Storage.GenPreSignUrl(
				ctx,
				ossPath,
				ossstore.MarkDown,
				ossstore.Get,
				20*time.Minute,
			)
			if err != nil {
				return nil, "", err
			} else {
				preUrl = presign.URL
			}

			// 将生成的URL存入缓存，过期时间20分钟
			err = storage.Storage.Cache.SetWithExpired(ctx, storage.BuildBlogCacheKey(blogDto.BlogId), preUrl, 20*time.Minute)
			if err != nil {
				msg := fmt.Sprintf("缓存预签名URL失败: %v", err)
				logger.Warn(msg)
				return nil, "", errors.New(msg)
			}
		}
	} else {
		// 博客不存在，记录警告日志并返回错误
		msg := fmt.Sprintf("博客不存在，id: %s", id)
		logger.Warn(msg)
		return nil, "", errors.New(msg)
	}

	// 返回博客视图对象和预签名URL
	return blogVo, preUrl, nil
}

// GetDisplayedFriendLinks 获取所有显示状态为 true 的友链信息
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//
// 返回值:
//   - []*dto.FriendLinkDto: 包含友链信息的 DTO 列表。
//   - error: 如果在查询友链时发生错误，则返回该错误。
func GetDisplayedFriendLinks(ctx context.Context) ([]*dto.FriendLinkDto, error) {
	return friendlinkrepo.FindDisplayedFriendLinks(ctx)
}

// ApplyFriendLink 申请友链
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//   - friendLinkDto: 友链申请信息。
//
// 返回值:
//   - error: 如果申请过程中发生错误，则返回该错误。
func ApplyFriendLink(ctx context.Context, friendLinkDto *dto.FriendLinkDto) error {
	// 1. 验证必需字段
	if friendLinkDto.FriendLinkName == "" {
		return errors.New("友链名称不能为空")
	}
	if friendLinkDto.FriendLinkUrl == "" {
		return errors.New("友链URL不能为空")
	}

	// 2. 检查该URL是否已存在
	existingLink, err := friendlinkrepo.FindFriendLinkByUrl(ctx, friendLinkDto.FriendLinkUrl)
	if err != nil {
		return fmt.Errorf("检查友链URL失败: %w", err)
	}
	if existingLink != nil {
		return errors.New("该友链URL已存在")
	}

	// 3. 设置申请友链为不显示状态
	friendLinkDto.Display = false

	// 4. 开启数据库事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 5. 创建友链记录
	if err := friendlinkrepo.CreateFriendLink(tx, friendLinkDto); err != nil {
		tx.Rollback()
		return fmt.Errorf("创建友链失败: %w", err)
	}

	// 6. 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	logger.Info(fmt.Sprintf("友链申请成功: %s - %s", friendLinkDto.FriendLinkName, friendLinkDto.FriendLinkUrl))
	return nil
}

// GetCommentsByBlogId 根据博客ID获取评论（业务端功能）
// - ctx: 上下文对象
// - blogId: 博客ID
//
// 返回值:
// - []vo.CommentVo: 评论列表
// - error: 错误信息
func GetCommentsByBlogId(ctx context.Context, blogId string) ([]vo.CommentVo, error) {
	// 获取楼主评论
	commentDtos, err := commentrepo.FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		return nil, err
	}

	// 根据博客ID查询博客标题（只查询一次，因为都是同一篇博客的评论）
	blogTitle, err := blogrepo.FindBlogTitleById(ctx, blogId)
	if err != nil {
		logger.Warn("查询博客标题失败，BlogId: %s, 错误: %v", blogId, err)
		blogTitle = "" // 设置为空字符串
	}

	// 保存所有楼主评论
	var commentVos []vo.CommentVo

	// 遍历所有楼主评论
	for _, commentDto := range commentDtos {
		// 创建楼主评论Vo
		commentVo := vo.CommentVo{
			CommentId:        commentDto.CommentId,
			CommenterEmail:   commentDto.CommenterEmail,
			BlogId:           commentDto.BlogId,
			BlogTitle:        blogTitle,
			OriginPostId:     commentDto.OriginPostId,
			ReplyToCommenter: commentDto.ReplyToCommenter,
			Content:          commentDto.Content,
			CreateTime:       commentDto.CreateTime,
		}

		// 获取楼层子评论
		subCommentDtos, err := commentrepo.FindCommentsByOriginPostId(ctx, commentDto.CommentId)
		if err != nil {
			return nil, err
		}

		// 将子评论转为 Vo，并保存
		for _, subCommentDto := range subCommentDtos {
			commentVo.SubComments = append(commentVo.SubComments, vo.CommentVo{
				CommentId:        subCommentDto.CommentId,
				CommenterEmail:   subCommentDto.CommenterEmail,
				BlogId:           subCommentDto.BlogId,
				BlogTitle:        blogTitle,
				OriginPostId:     subCommentDto.OriginPostId,
				ReplyToCommenter: subCommentDto.ReplyToCommenter,
				Content:          subCommentDto.Content,
				CreateTime:       subCommentDto.CreateTime,
			})
		}

		// 添加到楼主评论集合
		commentVos = append(commentVos, commentVo)
	}

	return commentVos, nil
}

// AddComment 添加评论（业务端功能）
// - ctx: 上下文对象
// - commentDto: 评论数据传输对象
//
// 返回值:
// - *vo.CommentVo: 创建的评论视图对象
// - error: 错误信息
func AddComment(ctx context.Context, commentDto *dto.CommentDto) (*vo.CommentVo, error) {
	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("添加评论事务失败: %v", r)
			tx.Rollback()
		}
	}()

	// 处理回复逻辑
	if commentDto.ReplyToCommentId != "" {
		// 如果是回复评论，需要查找被回复的评论信息
		replyToComment, err := commentrepo.FindCommentById(ctx, commentDto.ReplyToCommentId)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("被回复的评论不存在: %v", err)
		}

		// 如果回复的是楼主评论，则 OriginPostId 设置为被回复评论的ID
		// 如果回复的是子评论，则 OriginPostId 设置为原楼主评论的ID
		if replyToComment.OriginPostId == "" {
			// 回复的是楼主评论
			commentDto.OriginPostId = replyToComment.CommentId
		} else {
			// 回复的是子评论，保持原楼主评论ID
			commentDto.OriginPostId = replyToComment.OriginPostId
		}
	}

	// 保存到数据库
	resultDto, err := commentrepo.CreateComment(tx, commentDto)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("添加评论失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交添加评论事务失败: %v", err)
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 根据博客ID查询博客标题
	blogTitle, err := blogrepo.FindBlogTitleById(ctx, resultDto.BlogId)
	if err != nil {
		logger.Warn("查询博客标题失败，BlogId: %s, 错误: %v", resultDto.BlogId, err)
		blogTitle = "" // 设置为空字符串
	}

	// 转换为VO对象返回
	commentVo := &vo.CommentVo{
		CommentId:        resultDto.CommentId,
		CommenterEmail:   resultDto.CommenterEmail,
		BlogId:           resultDto.BlogId,
		BlogTitle:        blogTitle,
		OriginPostId:     resultDto.OriginPostId,
		ReplyToCommenter: resultDto.ReplyToCommenter,
		Content:          resultDto.Content,
		CreateTime:       resultDto.CreateTime,
	}

	return commentVo, nil
}

// GetLatestComments 获取最新的5条评论（业务端功能）
// - ctx: 上下文对象
//
// 返回值:
// - []vo.CommentVo: 最新的评论列表
// - error: 错误信息
func GetLatestComments(ctx context.Context) ([]vo.CommentVo, error) {
	// 获取最新的5条评论
	commentDtos, err := commentrepo.FindLatestComments(ctx, 5)
	if err != nil {
		return nil, err
	}

	var commentVos []vo.CommentVo

	// 将DTO转换为VO
	for _, commentDto := range commentDtos {
		// 根据博客ID查询博客标题
		blogTitle, err := blogrepo.FindBlogTitleById(ctx, commentDto.BlogId)
		if err != nil {
			logger.Warn("查询博客标题失败，BlogId: %s, 错误: %v", commentDto.BlogId, err)
			blogTitle = "" // 设置为空字符串
		}

		commentVo := vo.CommentVo{
			CommentId:        commentDto.CommentId,
			CommenterEmail:   commentDto.CommenterEmail,
			BlogId:           commentDto.BlogId,
			BlogTitle:        blogTitle,
			OriginPostId:     commentDto.OriginPostId,
			ReplyToCommenter: commentDto.ReplyToCommenter,
			Content:          commentDto.Content,
			CreateTime:       commentDto.CreateTime,
		}

		commentVos = append(commentVos, commentVo)
	}

	return commentVos, nil
}
