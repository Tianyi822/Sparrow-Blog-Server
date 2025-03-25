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

// DeleteBlogById 根据博客ID删除对应的博客数据。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - id: 要删除的博客的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteBlogById(ctx context.Context, id string) error {
	// 开启数据库事务，并绑定上下文以确保事务与请求生命周期一致
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 捕获可能的 panic，记录错误日志并回滚事务，避免事务未正确关闭。
		if r := recover(); r != nil {
			logger.Error("删除博客数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("删除 ID 为 %v 的博客记录", id)
	// 根据博客ID删除对应的博客记录
	if err := tx.Where("b_id = ?", id).Delete(&po.Blog{}).Error; err != nil {
		// 如果删除操作失败，回滚事务并记录错误日志。
		tx.Rollback()
		msg := fmt.Sprintf("删除博客数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 提交事务，确保事务被正确关闭
	tx.Commit()
	logger.Info("成功删除 ID 为 %v 的博客记录", id)

	// 如果删除成功，事务会自动提交（defer 中不会触发 Rollback）。
	return nil
}
