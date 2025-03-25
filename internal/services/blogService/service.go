package blogService

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repository/blogRepo"
	"h2blog_server/internal/repository/categoryRepo"
	"h2blog_server/internal/repository/tagRepo"
)

// UpdateBlogData 更新博客数据，包括分类、标签以及博客与标签的关联关系。
// 参数:
// - ctx: 上下文对象，用于控制请求生命周期和传递上下文信息。
// - blogDto: 包含博客数据的 DTO 对象，包含博客的基本信息、分类和标签。
//
// 返回值:
// - error: 如果在更新过程中发生错误，则返回具体的错误信息；否则返回 nil。
func UpdateBlogData(ctx context.Context, blogDto *dto.BlogDto) error {
	// 如果 blogDto 中没有 CategoryId，则表示该分类是新的，需要新建分类。
	if len(blogDto.CategoryId) == 0 {
		categoryDto := dto.CategoryDto{
			CName: blogDto.Category.CName,
		}
		err := categoryRepo.AddCategory(ctx, &categoryDto)
		if err != nil {
			return err
		}
		blogDto.CategoryId = categoryDto.CId
	}

	// 检查 blogDto 中的标签 ID，如果标签没有 ID，则需要创建新的标签。
	if len(blogDto.Tags) != 0 {
		// 分离出有 ID 的标签和没有 ID 的标签。
		var tagsWithId []dto.TagDto
		var tagsWithoutId []dto.TagDto

		for _, tag := range blogDto.Tags {
			if tag.TId != "" {
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

	// 建立标签与博客的关联关系。
	if err := tagRepo.AddBlogTagAssociation(ctx, blogDto.BId, blogDto.Tags); err != nil {
		return err
	}

	// 调用仓库方法保存更新后的博客数据。
	if err := blogRepo.UpdateBlog(ctx, blogDto); err != nil {
		return err
	}

	return nil
}

func GetBlogsInPage(ctx context.Context, page, pageSize int) ([]*dto.BlogDto, error) {
	blogDtos, err := blogRepo.FindBlogsInPage(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 获取 Tag 数据
	for _, blogDto := range blogDtos {
		tags, err := tagRepo.FindTagsByBlogId(ctx, blogDto.BId)
		if err != nil {
			return nil, err
		}

		tagDtos := make([]dto.TagDto, len(tags))
		for i, tag := range tags {
			tagDtos[i] = dto.TagDto{
				TName: tag.TName,
			}
		}

		blogDto.Tags = tagDtos
	}

	// 获取分类数据
	for _, blogDto := range blogDtos {
		category, err := categoryRepo.FindCategoryById(ctx, blogDto.CategoryId)
		if err != nil {
			return nil, err
		}
		blogDto.Category = dto.CategoryDto{
			CName: category.CName,
		}
	}

	return blogDtos, nil
}
