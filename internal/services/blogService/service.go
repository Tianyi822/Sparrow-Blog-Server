package blogService

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repository/blogRepo"
	"h2blog_server/internal/repository/categoryRepo"
	"h2blog_server/internal/repository/tagRepo"
)

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
