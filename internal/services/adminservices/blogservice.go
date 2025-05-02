package adminservices

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repositories/blogrepo"
	"h2blog_server/internal/repositories/categoryrepo"
	"h2blog_server/internal/repositories/tagrepo"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
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

	// 删除博客对应的 Markdown 文件
	err = storage.Storage.DeleteObject(ctx, ossstore.GenOssSavePath(blogTitle, ossstore.MarkDown))
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	tx.Commit()
	logger.Info("删除博客数据成功")

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
		if err := blogrepo.UpdateBlog(tx, blogDto); err != nil {
			tx.Rollback()
			return err
		}

		// 更新标签与博客的关联关系
		if err := tagrepo.UpdateBlogTagAssociation(tx, blogDto.BlogId, blogDto.Tags); err != nil {
			tx.Rollback()
			return err
		}
	}
	logger.Info("完成博客的更新或创建操作")

	// 提交事务
	tx.Commit()

	return nil
}
