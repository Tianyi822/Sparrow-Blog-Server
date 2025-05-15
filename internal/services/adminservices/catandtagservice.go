package adminservices

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/repositories/categoryrepo"
	"sparrow_blog_server/internal/repositories/tagrepo"
)

// GetAllCategoriesAndTags 获取所有的分类和标签信息。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递上下文信息。
//
// 返回值:
//   - []*dto.CategoryDto: 包含所有分类信息的 DTO 列表。
//   - []*dto.TagDto: 包含所有标签信息的 DTO 列表。
//   - error: 如果在获取分类或标签时发生错误，则返回具体的错误信息；否则返回 nil。
func GetAllCategoriesAndTags(ctx context.Context) ([]*dto.CategoryDto, []*dto.TagDto, error) {
	categories, err := categoryrepo.FindAllCategories(ctx)
	if err != nil {
		return nil, nil, err
	}

	tags, err := tagrepo.FindAllTags(ctx)
	if err != nil {
		return nil, nil, err
	}

	return categories, tags, nil
}
