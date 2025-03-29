package adminService

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repository/blogRepo"
	"h2blog_server/internal/repository/categoryRepo"
	"h2blog_server/internal/repository/tagRepo"
)

// UpdateOrAddBlog 更新或添加博客信息，并处理相关的分类和标签逻辑。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//   - blogDto: 包含博客信息的数据传输对象，包括博客内容、分类和标签等信息。
//
// 返回值:
//   - error: 如果操作过程中发生错误，则返回具体的错误信息；否则返回 nil。
func UpdateOrAddBlog(ctx context.Context, blogDto *dto.BlogDto) error {
	// 如果 blogDto 中没有 CategoryId，则表示该分类是新的，需要新建分类。
	if len(blogDto.CategoryId) == 0 {
		categoryDto := dto.CategoryDto{
			CategoryName: blogDto.Category.CategoryName,
		}
		err := categoryRepo.AddCategory(ctx, &categoryDto)
		if err != nil {
			return err
		}
		blogDto.CategoryId = categoryDto.CategoryId
	}

	// 检查 blogDto 中的标签 ID，如果标签没有 ID，则需要创建新的标签。
	if len(blogDto.Tags) != 0 {
		// 分离出有 ID 的标签和没有 ID 的标签。
		var tagsWithId []dto.TagDto
		var tagsWithoutId []dto.TagDto

		for _, tag := range blogDto.Tags {
			if tag.TagId != "" {
				tagsWithId = append(tagsWithId, tag)
			} else {
				tagsWithoutId = append(tagsWithoutId, tag)
			}
		}

		// 如果存在没有 ID 的标签，则调用仓库方法批量创建新标签。
		if len(tagsWithoutId) != 0 {
			newTags, err := tagRepo.AddTags(ctx, tagsWithoutId)
			if err != nil {
				return err
			}
			// 将有 ID 的标签和新创建的标签合并回 blogDto。
			blogDto.Tags = append(tagsWithId, newTags...)
		}
	}

	// 根据 blogDto 是否包含 BlogId 判断是新增博客还是更新博客。
	if len(blogDto.BlogId) == 0 {
		if err := blogRepo.AddBlog(ctx, blogDto); err != nil {
			return err
		}

		// 建立标签与博客的关联关系
		if err := tagRepo.AddBlogTagAssociation(ctx, blogDto.BlogId, blogDto.Tags); err != nil {
			return err
		}
	} else {
		if err := blogRepo.UpdateBlog(ctx, blogDto); err != nil {
			return err
		}

		// 更新标签与博客的关联关系
		if err := tagRepo.UpdateBlogTagAssociation(ctx, blogDto.BlogId, blogDto.Tags); err != nil {
			return err
		}
	}

	return nil
}

// GetBlogsToAdminPosts 获取所有博客及其关联的标签和分类信息。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//
// 返回值:
//   - []*dto.BlogDto: 包含博客及其关联标签和分类信息的 DTO 列表。
//   - error: 如果在查询博客、标签或分类时发生错误，则返回该错误。
func GetBlogsToAdminPosts(ctx context.Context) ([]*dto.BlogDto, error) {
	blogDtos, err := blogRepo.FindAllBlogs(ctx, false)
	if err != nil {
		return nil, err
	}

	// 遍历博客列表，为每个博客获取其关联的标签数据。
	for _, blogDto := range blogDtos {
		tags, err := tagRepo.FindTagsByBlogId(ctx, blogDto.BlogId)
		if err != nil {
			return nil, err
		}

		blogDto.Tags = tags
	}

	// 遍历博客列表，为每个博客获取其关联的分类数据。
	for _, blogDto := range blogDtos {
		category, err := categoryRepo.FindCategoryById(ctx, blogDto.CategoryId)
		if err != nil {
			return nil, err
		}
		blogDto.Category = dto.CategoryDto{
			CategoryId:   category.CategoryId,
			CategoryName: category.CategoryName,
		}
	}

	return blogDtos, nil
}

// DeleteBlog 删除指定ID的博客，并根据相关联的数据进行清理操作。
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - id: 要删除的博客的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteBlog(ctx context.Context, id string) error {
	// 获取与博客关联的分类ID。
	categoryId, err := blogRepo.GetCategoryIdByBlogId(ctx, id)
	if err != nil {
		return err
	}

	// 获取与博客关联的所有标签。
	tags, err := tagRepo.FindTagsByBlogId(ctx, id)
	if err != nil {
		return err
	}

	// 调用仓库方法根据ID删除博客。
	err = blogRepo.DeleteBlogById(ctx, id)
	if err != nil {
		return err
	}

	// 统计该分类下剩余的博客数量。
	num, err := blogRepo.CalBlogsCountByCategoryId(ctx, categoryId)
	if err != nil {
		return err
	}

	// 如果该分类下没有博客，则删除该分类。
	if num == 0 {
		err = categoryRepo.DeleteCategoryById(ctx, categoryId)
		if err != nil {
			return err
		}
	}

	// 遍历所有与博客关联的标签，检查每个标签是否还有其他博客关联。
	for _, tag := range tags {
		num, err = tagRepo.CalBlogsCountByTagId(ctx, tag.TagId)
		if err != nil {
			return err
		}

		// 如果某个标签没有其他博客关联，则删除该标签。
		if num == 0 {
			err = tagRepo.DeleteTagById(ctx, tag.TagId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetAllCategoriesAndTags 获取所有的分类和标签信息。
//
// 参数:
// - ctx: 上下文对象，用于控制请求生命周期和传递上下文信息。
//
// 返回值:
// - []*dto.CategoryDto: 包含所有分类信息的 DTO 列表。
// - []*dto.TagDto: 包含所有标签信息的 DTO 列表。
// - error: 如果在获取分类或标签时发生错误，则返回具体的错误信息；否则返回 nil。
func GetAllCategoriesAndTags(ctx context.Context) ([]*dto.CategoryDto, []*dto.TagDto, error) {
	categories, err := categoryRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, nil, err
	}

	tags, err := tagRepo.GetAllTags(ctx)
	if err != nil {
		return nil, nil, err
	}

	return categories, tags, nil
}

func SetTop(ctx context.Context, id string) error {
	err := blogRepo.SetTopById(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func ChangeBlogState(ctx context.Context, id string) error {
	err := blogRepo.ChangeBlogStateById(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
