package commentrepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
	"strings"
)

// FindCommentsByContentLike 根据评论内容模糊查询评论
// - ctx: 上下文对象
// - content: 评论内容
//
// 返回值:
// - []po.Comment: 符合模糊查询的评论列表
// - error: 错误信息
func FindCommentsByContentLike(ctx context.Context, content string) ([]po.Comment, error) {
	var comments []po.Comment

	logger.Info("查询评论数据")
	result := storage.Storage.Db.Model(&po.Comment{}).
		WithContext(ctx).
		Where("LOWER(content) LIKE ?", "%"+strings.ToLower(content)+"%").
		Find(&comments)
	if result.Error != nil {
		msg := fmt.Sprintf("查询评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	logger.Info("查询评论数据成功: %v", result.RowsAffected)

	return comments, nil
}

// FindCommentsByBlogId 根据博客ID查询评论
// - ctx: 上下文对象
// - blogId: 博客ID
//
// 返回值:
// - []po.Comment: 符合博客ID的评论列表
// - error: 错误信息
func FindCommentsByBlogId(ctx context.Context, blogId string) ([]po.Comment, error) {
	var comments []po.Comment

	logger.Info("根据博客 ID 查询评论数据")
	result := storage.Storage.Db.Model(&po.Comment{}).
		WithContext(ctx).
		Where("blog_id = ?", blogId).
		Find(&comments)
	if result.Error != nil {
		msg := fmt.Sprintf("根据博客 ID 查询评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	logger.Info("根据博客 ID 查询评论数据成功: %v", result.RowsAffected)

	return comments, nil
}

// FindCommentsByOriginPostId 根据楼主评论ID查询评论
// - ctx: 上下文对象
// - parentId: 楼主评论ID
//
// 返回值:
// - []po.Comment: 符合楼主评论ID的评论列表
// - error: 错误信息
func FindCommentsByOriginPostId(ctx context.Context, originPostId string) ([]po.Comment, error) {
	var comments []po.Comment

	logger.Info("根据父评论 ID 查询评论数据")
	result := storage.Storage.Db.Model(&po.Comment{}).
		WithContext(ctx).
		Where("original_poster_id = ?", originPostId).
		Find(&comments)
	if result.Error != nil {
		msg := fmt.Sprintf("根据父评论 ID 查询评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	logger.Info("根据父评论 ID 查询评论数据成功: %v", result.RowsAffected)

	return comments, nil
}

// AddComment 创建新的评论
// - ctx: 上下文对象
// - po: 评论实体指针
//
// 返回值:
// - int64: 受影响的行数
// - error: 错误信息
func AddComment(ctx context.Context, po *po.Comment) (int64, error) {
	// 开启事务
	tx := storage.Storage.Db.Model(po).WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("创建评论数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("添加评论数据")
	// 执行创建操作
	result := tx.Create(po)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("创建评论数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("创建评论数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}

// DeleteCommentById 根据评论ID删除评论
// - ctx: 上下文对象
// - id: 评论ID
//
// 返回值:
// - int64: 受影响的行数
// - error: 错误信息
func DeleteCommentById(ctx context.Context, id string) (int64, error) {
	tx := storage.Storage.Db.Model(&po.Comment{}).WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除评论数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("删除评论数据")
	result := tx.Delete(&po.Comment{CommentId: id})
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("删除评论数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("删除评论数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}
