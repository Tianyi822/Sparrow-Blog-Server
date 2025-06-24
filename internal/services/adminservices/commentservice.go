package adminservices

import (
	"context"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/vo"
	"sparrow_blog_server/internal/repositories/blogrepo"
	"sparrow_blog_server/internal/repositories/commentrepo"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
)

// UpdateComment 更新评论（管理员功能）
// - ctx: 上下文对象
// - commentId: 评论ID
// - commentDto: 评论数据传输对象
//
// 返回值:
// - *vo.CommentVo: 更新后的评论视图对象
// - error: 错误信息
func UpdateComment(ctx context.Context, commentId string, commentDto *dto.CommentDto) (*vo.CommentVo, error) {
	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新评论事务失败: %v", r)
			tx.Rollback()
		}
	}()

	// 检查评论是否存在
	existingCommentDto, err := commentrepo.FindCommentById(ctx, commentId)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("评论不存在: %v", err)
	}

	// 更新评论内容
	existingCommentDto.Content = commentDto.Content

	// 保存更新
	updatedDto, err := commentrepo.UpdateComment(ctx, tx, existingCommentDto)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新评论失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交更新评论事务失败: %v", err)
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 根据博客ID查询博客标题
	blogTitle, err := blogrepo.FindBlogTitleById(ctx, updatedDto.BlogId)
	if err != nil {
		logger.Warn("查询博客标题失败，BlogId: %s, 错误: %v", updatedDto.BlogId, err)
		blogTitle = "" // 设置为空字符串
	}

	// 转换为VO对象返回
	commentVo := &vo.CommentVo{
		CommentId:        updatedDto.CommentId,
		CommenterEmail:   updatedDto.CommenterEmail,
		BlogTitle:        blogTitle,
		OriginPostId:     updatedDto.OriginPostId,
		ReplyToCommentId: updatedDto.ReplyToCommentId,
		Content:          updatedDto.Content,
		CreateTime:       updatedDto.CreateTime,
	}

	return commentVo, nil
}

// DeleteComment 删除评论（管理员功能）
// - ctx: 上下文对象
// - commentId: 评论ID
//
// 返回值:
// - error: 错误信息
func DeleteComment(ctx context.Context, commentId string) error {
	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除评论事务失败: %v", r)
			tx.Rollback()
		}
	}()

	// 检查评论是否存在
	_, err := commentrepo.FindCommentById(ctx, commentId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("评论不存在: %v", err)
	}

	// 删除评论
	_, err = commentrepo.DeleteCommentById(ctx, tx, commentId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除评论失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交删除评论事务失败: %v", err)
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// DeleteCommentWithSubComments 删除评论及其所有子评论（管理员功能）
// - ctx: 上下文对象
// - commentId: 评论ID
//
// 返回值:
// - error: 错误信息
func DeleteCommentWithSubComments(ctx context.Context, commentId string) error {
	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除评论及子评论事务失败: %v", r)
			tx.Rollback()
		}
	}()

	// 检查评论是否存在
	comment, err := commentrepo.FindCommentById(ctx, commentId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("评论不存在: %v", err)
	}

	// 如果是主评论（OriginPostId为空），需要删除所有子评论
	if comment.OriginPostId == "" {
		// 这是主评论，删除所有子评论
		subComments, err := commentrepo.FindCommentsByOriginPostId(ctx, commentId)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("查询子评论失败: %v", err)
		}

		// 删除所有子评论
		for _, subComment := range subComments {
			_, err = commentrepo.DeleteCommentById(ctx, tx, subComment.CommentId)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("删除子评论失败: %v", err)
			}
		}
	}

	// 删除主评论本身
	_, err = commentrepo.DeleteCommentById(ctx, tx, commentId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除评论失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交删除评论事务失败: %v", err)
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetAllComments 获取所有评论（管理员功能）
// - ctx: 上下文对象
//
// 返回值:
// - []vo.CommentVo: 评论列表
// - error: 错误信息
func GetAllComments(ctx context.Context) ([]vo.CommentVo, error) {
	// 获取所有评论
	commentDtos, err := commentrepo.FindAllComments(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为VO对象
	var commentVos []vo.CommentVo
	for _, commentDto := range commentDtos {
		// 根据博客ID查询博客标题
		blogTitle, err := blogrepo.FindBlogTitleById(ctx, commentDto.BlogId)
		if err != nil {
			// 如果查询博客标题失败，记录警告但继续处理其他评论
			logger.Warn("查询博客标题失败，BlogId: %s, 错误: %v", commentDto.BlogId, err)
			blogTitle = "" // 设置为空字符串
		}

		commentVo := vo.CommentVo{
			CommentId:        commentDto.CommentId,
			CommenterEmail:   commentDto.CommenterEmail,
			BlogTitle:        blogTitle,
			OriginPostId:     commentDto.OriginPostId,
			ReplyToCommentId: commentDto.ReplyToCommentId,
			Content:          commentDto.Content,
			CreateTime:       commentDto.CreateTime,
		}
		commentVos = append(commentVos, commentVo)
	}

	return commentVos, nil
}
