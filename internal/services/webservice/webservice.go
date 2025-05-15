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
