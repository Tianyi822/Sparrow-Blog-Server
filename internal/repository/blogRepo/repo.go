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

func FindBlogsInPage(ctx context.Context, page, pageSize int) ([]*dto.BlogDto, error) {
	blogs := make([]*po.Blog, 0)

	// 查询博客信息数据
	result := storage.Storage.Db.Model(&po.Blog{}).
		WithContext(ctx).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&blogs).
		Order("H2_BLOG_INFO.create_time DESC")

	if result.Error != nil {
		msg := fmt.Sprintf("查询博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	blogDtos := make([]*dto.BlogDto, 0)

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

	return blogDtos, nil
}
