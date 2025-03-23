package blogInfoRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
)

// FindBlogById 根据博客ID查询单条博客信息
// - ctx: 上下文对象
// - id: 博客ID
func FindBlogById(ctx context.Context, id string) (*po.HBlog, error) {
	blog := &po.HBlog{}

	// 查询博客信息数据
	result := storage.Storage.Db.Model(blog).WithContext(ctx).Where("H2_BLOG_INFO.blog_id = ?", id).First(blog)

	// 数据不存在或者发生错误
	if result.Error != nil {
		msg := fmt.Sprintf("查询博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 查询成功
	return blog, nil
}
