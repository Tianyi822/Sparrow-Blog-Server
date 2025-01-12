package po

import (
	"context"
	"errors"
	"fmt"
	"h2blog/pkg/logger"
	"h2blog/storage"
	"time"
)

type BlogInfo struct {
	BlogId     string    `gorm:"column:blog_id;primaryKey;"`                                  // 博客ID
	Title      string    `gorm:"column:title;unique"`                                         // 博客标题
	Brief      string    `gorm:"column:brief"`                                                // 博客简介
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (hbi *BlogInfo) TableName() string {
	return "H2_BLOG_INFO"
}

// FindOneById 根据博客ID查找博客信息
// 参数:
//   - ctx: 上下文对象，用于控制请求的上下文，如超时、取消等
//
// 返回值:
//   - error: 如果查找过程中发生错误，返回错误信息；否则返回nil
func (hbi *BlogInfo) FindOneById(ctx context.Context) (int64, error) {
	// 查询博客信息数据
	result := storage.Storage.Db.WithContext(ctx).Model(hbi).Where("H2_BLOG_INFO.blog_id = ?", hbi.BlogId).First(hbi)

	// 检查查询结果是否有错误
	if result.Error != nil {
		// 如果查询过程中发生其他错误，则记录错误日志并返回错误
		msg := fmt.Sprintf("查询博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return result.RowsAffected, errors.New(msg)
	}
	// 如果查询成功，返回nil表示没有错误
	return result.RowsAffected, nil
}

// UpdateOneById 根据博客ID更新博客信息数据
// 参数:
//
//	ctx: 上下文对象，用于控制请求的生命周期
//
// 返回值:
//
//	affectedNum: 受影响的行数
//	err: 错误信息，如果操作成功则为nil
func (hbi *BlogInfo) UpdateOneById(ctx context.Context) (affectedNum int64, err error) {
	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 恢复机制，用于捕获和处理 panic
		if r := recover(); r != nil {
			// 如果发生 panic，记录错误日志并回滚事务
			logger.Error("更新博客信息数据失败: %v", r)
			tx.Rollback()
		} else if err != nil {
			// 如果有其他错误，回滚事务
			tx.Rollback()
		} else {
			// 如果没有错误，提交事务
			tx.Commit()
		}
	}()

	logger.Info("更新博客信息数据")

	// 使用事务更新博客信息数据
	result := tx.Model(hbi).Where("H2_BLOG_INFO.blog_id = ?", hbi.BlogId).Updates(hbi)
	if result.Error != nil {
		// 如果更新数据失败，记录错误日志并返回错误
		msg := fmt.Sprintf("更新博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}
	// 记录更新成功的日志并返回受影响的行数
	logger.Info("更新博客信息数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}

// AddOne 添加一条博客信息数据
func (hbi *BlogInfo) AddOne(ctx context.Context) (affectedNum int64, err error) {
	//开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 恢复机制，用于捕获和处理 panic
		if r := recover(); r != nil {
			// 如果发生 panic，记录错误日志并回滚事务
			logger.Error("创建博客信息数据失败: %v", r)
			tx.Rollback()
		} else if err != nil {
			// 如果有其他错误，回滚事务
			tx.Rollback()
		} else {
			// 如果没有错误，提交事务
			tx.Commit()
		}
	}()

	// 如果查询失败，则添加新数据
	logger.Info("添加博客信息数据")

	// 使用事务创建博客信息数据
	result := tx.Create(hbi)
	if result.Error != nil {
		// 如果创建数据失败，记录错误日志并返回错误
		msg := fmt.Sprintf("创建博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 记录创建成功的日志并返回受影响的行数
	logger.Info("创建博客信息数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}

// DeleteOneById 根据博客ID删除一条博客信息数据
// 参数:
//
//	ctx context.Context: 上下文对象，用于控制请求的生命周期
//
// 返回值:
//
//	affectedNum int64: 受影响的行数
//	err error: 错误信息，如果操作成功则为nil
func (hbi *BlogInfo) DeleteOneById(ctx context.Context) (affectedNum int64, err error) {
	//开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 恢复机制，用于捕获和处理 panic
		if r := recover(); r != nil {
			// 如果发生 panic，记录错误日志并回滚事务
			logger.Error("创建博客信息数据失败: %v", r)
			tx.Rollback()
		} else if err != nil {
			// 如果有其他错误，回滚事务
			tx.Rollback()
		} else {
			// 如果没有错误，提交事务
			tx.Commit()
		}
	}()

	logger.Info("删除博客信息数据")

	// 使用事务删除博客信息数据
	result := tx.Model(hbi).Where("H2_BLOG_INFO.blog_id = ?", hbi.BlogId).Delete(hbi)
	if result.Error != nil {
		// 如果删除数据失败，记录错误日志并返回错误
		msg := fmt.Sprintf("删除博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}
	// 记录删除成功的日志并返回受影响的行数
	logger.Info("删除博客信息数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}
