package commentservice

import (
	"context"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/vo"
	"sparrow_blog_server/internal/repositories/commentrepo"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
)

// GetCommentsByBlogId 根据博客ID获取评论
// - ctx: 上下文对象
// - blogId: 博客ID
//
// 返回值:
// - []vo.CommentVo: 评论列表
// - error: 错误信息
func GetCommentsByBlogId(ctx context.Context, blogId string) ([]vo.CommentVo, error) {
	// 获取楼主评论
	commentDtos, err := commentrepo.FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		return nil, err
	}

	// 保存所有楼主评论
	var commentVos []vo.CommentVo

	// 遍历所有楼主评论
	for _, commentDto := range commentDtos {
		// 创建楼主评论Vo
		commentVo := vo.CommentVo{
			CommentId:      commentDto.CommentId,
			CommenterEmail: commentDto.CommenterEmail,
			BlogId:         commentDto.BlogId,
			OriginPostId:   commentDto.OriginPostId,
			Content:        commentDto.Content,
			CreateTime:     commentDto.CreateTime,
		}

		// 获取楼层子评论
		subCommentDtos, err := commentrepo.FindCommentsByOriginPostId(ctx, commentDto.CommentId)
		if err != nil {
			return nil, err
		}

		// 将子评论转为 Vo，并保存
		for _, subCommentDto := range subCommentDtos {
			commentVo.SubComments = append(commentVo.SubComments, vo.CommentVo{
				CommentId:      subCommentDto.CommentId,
				CommenterEmail: subCommentDto.CommenterEmail,
				BlogId:         subCommentDto.BlogId,
				OriginPostId:   subCommentDto.OriginPostId,
				Content:        subCommentDto.Content,
				CreateTime:     subCommentDto.CreateTime,
			})
		}

		// 添加到楼主评论集合
		commentVos = append(commentVos, commentVo)
	}

	return commentVos, nil
}

// AddComment 添加评论
// - ctx: 上下文对象
// - commentDto: 评论数据传输对象
//
// 返回值:
// - *vo.CommentVo: 创建的评论视图对象
// - error: 错误信息
func AddComment(ctx context.Context, commentDto *dto.CommentDto) (*vo.CommentVo, error) {
	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("添加评论事务失败: %v", r)
			tx.Rollback()
		}
	}()

	// 保存到数据库
	resultDto, err := commentrepo.CreateComment(ctx, tx, commentDto)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("添加评论失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交添加评论事务失败: %v", err)
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 转换为VO对象返回
	commentVo := &vo.CommentVo{
		CommentId:      resultDto.CommentId,
		CommenterEmail: resultDto.CommenterEmail,
		BlogId:         resultDto.BlogId,
		OriginPostId:   resultDto.OriginPostId,
		Content:        resultDto.Content,
		CreateTime:     resultDto.CreateTime,
	}

	return commentVo, nil
}

// UpdateComment 更新评论
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

	// 转换为VO对象返回
	commentVo := &vo.CommentVo{
		CommentId:      updatedDto.CommentId,
		CommenterEmail: updatedDto.CommenterEmail,
		BlogId:         updatedDto.BlogId,
		OriginPostId:   updatedDto.OriginPostId,
		Content:        updatedDto.Content,
		CreateTime:     updatedDto.CreateTime,
	}

	return commentVo, nil
}

// DeleteComment 删除评论
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
