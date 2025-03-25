package blogRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
)

// FindBlogsInPage 查询指定页码的博客列表，并返回分页后的博客数据。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和取消操作。
//   - page: 当前页码，从1开始计数。
//   - pageSize: 每页显示的博客数量。
//
// 返回值:
//   - []*dto.BlogDto: 包含博客信息的DTO（数据传输对象）列表。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回nil。
func FindBlogsInPage(ctx context.Context, page, pageSize int) ([]*dto.BlogDto, error) {
	blogs := make([]*po.Blog, 0)

	// 查询博客信息数据，按置顶优先和创建时间倒序排序，并进行分页处理。
	result := storage.Storage.Db.Model(&po.Blog{}).
		WithContext(ctx).
		Order("is_top DESC").
		Order("create_time DESC").
		Find(&blogs).
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	// 如果查询过程中发生错误，记录错误日志并返回错误信息。
	if result.Error != nil {
		msg := fmt.Sprintf("查询博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 初始化博客DTO列表，用于存储转换后的博客数据。
	blogDtos := make([]*dto.BlogDto, 0)

	// 遍历查询到的博客数据，将其转换为DTO格式并添加到结果列表中。
	for _, blog := range blogs {
		blogDto := &dto.BlogDto{
			BId:        blog.BId,
			Brief:      blog.Brief,
			Title:      blog.Title,
			IsTop:      blog.IsTop,
			State:      blog.State,
			WordsNum:   blog.WordsNum,
			CategoryId: blog.CategoryId,
			CreateTime: blog.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: blog.UpdateTime.Format("2006-01-02 15:04:05"),
		}
		blogDtos = append(blogDtos, blogDto)
	}

	// 返回转换后的博客DTO列表和nil错误。
	return blogDtos, nil
}
