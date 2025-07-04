package adminservices

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/repositories/blogrepo"
	"sparrow_blog_server/internal/repositories/categoryrepo"
	"sparrow_blog_server/internal/repositories/commentrepo"
	"sparrow_blog_server/internal/repositories/tagrepo"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/searchengine"
	"sparrow_blog_server/storage"
	"sparrow_blog_server/storage/ossstore"
	"time"
)

// GetBlogsToAdminPosts 获取所有博客及其关联的标签和分类信息。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//
// 返回值:
//   - []*dto.BlogDto: 包含博客及其关联标签和分类信息的 DTO 列表。
//   - error: 如果在查询博客、标签或分类时发生错误，则返回该错误。
func GetBlogsToAdminPosts(ctx context.Context) ([]*dto.BlogDto, error) {
	blogDtos, err := blogrepo.FindAllBlogs(ctx, false)
	if err != nil {
		return nil, err
	}

	// 遍历博客列表，为每个博客获取其关联的标签数据。
	for _, blogDto := range blogDtos {
		tags, err := tagrepo.FindTagsByBlogId(ctx, blogDto.BlogId)
		if err != nil {
			return nil, err
		}

		blogDto.Tags = tags
	}

	// 遍历博客列表，为每个博客获取其关联的分类数据。
	for _, blogDto := range blogDtos {
		category, err := categoryrepo.FindCategoryById(ctx, blogDto.CategoryId)
		if err != nil {
			return nil, err
		}
		blogDto.Category = category
	}

	return blogDtos, nil
}

// GetBlogData 根据博客ID获取博客的详细信息，包括关联的标签和分类。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - id: 要查询的博客的唯一标识符。
//
// 返回值:
//   - *dto.BlogDto: 包含博客详细信息的数据传输对象，包括标签和分类信息。
//   - string: 预签名 URL，用于读取 OSS 中的文章内容。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回 nil。
func GetBlogData(ctx context.Context, id string) (*dto.BlogDto, string, error) {
	// 根据博客ID从仓库中获取博客的基础信息。
	blogDto, err := blogrepo.FindBlogById(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// 根据博客ID获取与该博客关联的所有标签。
	tags, err := tagrepo.FindTagsByBlogId(ctx, blogDto.BlogId)
	if err != nil {
		return nil, "", err
	}
	blogDto.Tags = tags

	// 根据博客的分类ID获取分类信息，并将其映射为 CategoryDto 对象。
	category, err := categoryrepo.FindCategoryById(ctx, blogDto.CategoryId)
	if err != nil {
		return nil, "", err
	}
	blogDto.Category = category

	// 获取预签名 URL，用于读取 OSS 中文章内容
	presignUrl, err := storage.Storage.GenPreSignUrl(
		ctx,
		ossstore.GenOssSavePath(blogDto.BlogTitle, ossstore.MarkDown),
		ossstore.MarkDown,
		ossstore.Get,
		1*time.Minute,
	)
	if err != nil {
		return nil, "", err
	}

	// 返回包含完整信息的博客DTO对象。
	return blogDto, presignUrl.URL, nil
}

// DeleteBlogById 删除指定ID的博客
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - id: 要删除的博客的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteBlogById(ctx context.Context, id string) error {
	// 获取博客标题
	blogTitle, err := blogrepo.FindBlogTitleById(ctx, id)
	if err != nil {
		return err
	}

	// 开启删除博客事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除博客失败: %v", r)
			tx.Rollback()
		}
	}()

	// 调用仓库方法根据ID删除博客。
	err = blogrepo.DeleteBlogById(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除博客标签关联数据
	err = tagrepo.DeleteBlogTagAssociationByBlogId(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除博客相关的所有评论
	_, err = commentrepo.DeleteCommentsByBlogId(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除博客对应的 Markdown 文件
	err = storage.Storage.DeleteObject(ctx, ossstore.GenOssSavePath(blogTitle, ossstore.MarkDown))
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	tx.Commit()
	logger.Info("删除博客数据成功")

	// 从搜索索引中删除博客
	if err := searchengine.DeleteIndex(id); err != nil {
		logger.Warn("删除博客搜索索引失败: %v", err)
		// 注意：这里不返回错误，因为数据库操作已经成功，索引删除失败不应该影响整个删除操作
	}

	// 清理无用标签和分类
	cleanUpTx := storage.Storage.Db.WithContext(ctx).Begin()
	if err = tagrepo.CleanTagsWithoutBlog(cleanUpTx); err != nil {
		cleanUpTx.Rollback()
		return err
	}

	if err = categoryrepo.CleanCategoriesWithoutBlog(cleanUpTx); err != nil {
		cleanUpTx.Rollback()
		return err
	}
	cleanUpTx.Commit()

	return nil
}

func SetTop(ctx context.Context, id string) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("修改置顶状态失败: %v", r)
			tx.Rollback()
		}
	}()

	if err := blogrepo.SetTopById(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func ChangeBlogState(ctx context.Context, id string) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("修改博客状态失败: %v", r)
			tx.Rollback()
		}
	}()

	if err := blogrepo.ChangeBlogStateById(tx, id); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

// UpdateOrAddBlog 更新或添加博客信息，并处理相关的分类和标签逻辑。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//   - blogDto: 包含博客信息的数据传输对象，包括博客内容、分类和标签等信息。
//
// 返回值:
//   - error: 如果操作过程中发生错误，则返回具体的错误信息；否则返回 nil。
func UpdateOrAddBlog(ctx context.Context, blogDto *dto.BlogDto) error {
	// 记录是否为新增操作（在事务开始前判断）
	isNewBlog := len(blogDto.BlogId) == 0

	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新或创建博客数据失败: %v", r)
			tx.Rollback()
		}
	}()

	// 如果 blogDto 中没有 CategoryId，则表示该分类是新的，需要新建分类。
	if len(blogDto.CategoryId) == 0 {
		categoryDto := dto.CategoryDto{
			CategoryName: blogDto.Category.CategoryName,
		}
		err := categoryrepo.AddCategory(tx, &categoryDto)
		if err != nil {
			tx.Rollback()
			return err
		}
		blogDto.CategoryId = categoryDto.CategoryId
	}

	// 保存新建的标签
	var newTags []dto.TagDto

	// 检查 blogDto 中的标签 ID，如果标签没有 ID，则需要创建新的标签。
	if len(blogDto.Tags) != 0 {
		// 分离不携带 ID 的标签和携带 ID 的标签
		var tagsWithoutId []dto.TagDto
		var tagsWithId []dto.TagDto
		for _, tag := range blogDto.Tags {
			if tag.TagId != "" {
				tagsWithId = append(tagsWithId, tag)
			} else {
				tagsWithoutId = append(tagsWithoutId, tag)
			}
		}

		// 如果存在没有 ID 的标签，则调用仓库方法批量创建新标签。
		if len(tagsWithoutId) != 0 {
			var err error
			newTags, err = tagrepo.AddTags(tx, tagsWithoutId)
			if err != nil {
				tx.Rollback()
				return err
			}
			// 将有 ID 的标签和新创建的标签合并回 blogDto。
			blogDto.Tags = append(tagsWithId, newTags...)
		}
	}

	// 根据 blogDto 是否包含 BlogId 判断是新增博客还是更新博客。
	if len(blogDto.BlogId) == 0 {
		if err := blogrepo.AddBlog(tx, blogDto); err != nil {
			tx.Rollback()
			return err
		}

		// 建立标签与博客的关联关系
		if err := tagrepo.AddBlogTagAssociation(tx, blogDto.BlogId, blogDto.Tags); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 更新博客信息
		// 需要删除 OSS 中原有的文章，先从数据库中拿到原来的标题
		title, err := blogrepo.FindBlogTitleById(ctx, blogDto.BlogId)
		if err != nil {
			tx.Rollback()
			return err
		}

		if title != blogDto.BlogTitle {
			// 如果标题有变化，则需要删除 OSS 中的旧文章
			logger.Info("删除 OSS 中的旧文章: %s", title)
			if deleteErr := storage.Storage.DeleteObject(ctx, ossstore.GenOssSavePath(title, ossstore.MarkDown)); deleteErr != nil {
				logger.Warn("删除 OSS 中的旧文章失败: %v", deleteErr)
				tx.Rollback()
				return deleteErr
			}
			logger.Info("删除 OSS 中的旧文章成功")
		}

		// 再更新数据库元数据
		if updateErr := blogrepo.UpdateBlog(tx, blogDto); updateErr != nil {
			logger.Warn("更新博客数据失败: %v", updateErr)
			tx.Rollback()
			return updateErr
		}

		// 更新标签与博客的关联关系
		if updateTagErr := tagrepo.UpdateBlogTagAssociation(tx, blogDto.BlogId, blogDto.Tags); updateTagErr != nil {
			logger.Warn("更新标签与博客的关联关系失败: %v", updateTagErr)
			tx.Rollback()
			return updateTagErr
		}

		// 删除缓存中的博客预签名 URL
		if err = storage.Storage.Cache.Delete(ctx, storage.BuildBlogCacheKey(blogDto.BlogId)); err != nil {
			logger.Warn("删除缓存中的博客预签名 URL 失败: %v", err)
			tx.Rollback()
			return err
		}
	}
	logger.Info("完成博客的更新或创建操作")

	// 提交事务
	tx.Commit()

	// 将更新或者新增的博客添加到索引中
	// 注意：索引操作在事务提交后进行，确保数据库操作成功后再更新索引
	var indexErr error
	if isNewBlog {
		// 如果是新增操作，使用AddIndex
		indexErr = searchengine.AddIndex(ctx, blogDto)
	} else {
		// 如果是更新操作，使用UpdateIndex
		indexErr = searchengine.UpdateIndex(ctx, blogDto)
	}

	if indexErr != nil {
		logger.Warn("更新博客搜索索引失败: %v", indexErr)
		// 注意：这里不返回错误，因为数据库操作已经成功，索引更新失败不应该影响整个操作
	}

	return nil
}
