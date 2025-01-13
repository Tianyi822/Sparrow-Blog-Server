package blogInfoRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog/internal/model/po"
	"h2blog/pkg/logger"
	"h2blog/storage"
)

// FindBlogById 根据博客ID查询单条博客信息
// - ctx: 上下文对象
// - id: 博客ID
func FindBlogById(ctx context.Context, id string) (*po.BlogInfo, error) {
	blog := &po.BlogInfo{}

	// 查询博客信息数据
	result := storage.Storage.Db.Model(blog).WithContext(ctx).Where("H2_BLOG_INFO.blog_id = ?", id).First(blog)

	// 检查查询结果是否有错误
	if result.Error != nil {
		// 如果查询过程中发生其他错误，则记录错误日志并返回错误
		msg := fmt.Sprintf("查询博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 查询成功
	return blog, nil
}

// UpdateById 根据博客ID更新博客信息数据
// - ctx: 上下文对象
// - blogInfo: 博客信息
func UpdateById(ctx context.Context, blogInfo *po.BlogInfo) (int64, error) {
	tx := storage.Storage.Db.Model(blogInfo).WithContext(ctx).Begin()
	// 使用defer确保在panic时回滚事务
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新博客信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("更新博客信息数据")
	// 执行更新操作
	result := tx.Where("H2_BLOG_INFO.blog_id = ?", blogInfo.BlogId).Updates(blogInfo) // 执行更新操作
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("更新博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("更新博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 检查影响行数
	if result.RowsAffected == 0 {
		tx.Rollback()
		msg := "更新博客信息数据失败: 记录不存在或未发生变更"
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("更新博客信息数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}

// CreateBlogInfo 添加一条博客信息数据
// - ctx: 上下文对象
// - blogInfo: 博客信息
func CreateBlogInfo(ctx context.Context, blogInfo *po.BlogInfo) (int64, error) {
	tx := storage.Storage.Db.Model(blogInfo).WithContext(ctx).Begin()
	// 使用defer确保在panic时回滚事务
	defer func() {
		if r := recover(); r != nil {
			logger.Error("创建博客信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("添加博客信息数据")
	// 执行创建操作
	result := tx.Create(blogInfo)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("创建博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("创建博客信息数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}

// DeleteById 根据博客ID删除一条博客信息数据
// - ctx: 上下文对象
// - id: 博客ID
func DeleteById(ctx context.Context, id string) (int64, error) {
	// 开启事务
	tx := storage.Storage.Db.Model(&po.BlogInfo{}).WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除博客信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("删除博客信息数据")
	// 执行删除操作
	result := tx.Where("H2_BLOG_INFO.blog_id = ?", id).Delete(&po.BlogInfo{})
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("删除博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 检查影响行数
	if result.RowsAffected == 0 {
		tx.Rollback()
		msg := "删除博客信息数据失败: 记录不存在"
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("删除博客信息数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}
