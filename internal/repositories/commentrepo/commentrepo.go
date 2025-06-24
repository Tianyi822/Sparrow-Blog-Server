package commentrepo

import (
	"context"
	"errors"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/utils"
	"sparrow_blog_server/storage"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FindCommentsByContentLike 根据评论内容模糊查询评论
// - ctx: 上下文对象
// - content: 评论内容
//
// 返回值:
// - []dto.CommentDto: 符合模糊查询的评论列表
// - error: 错误信息
func FindCommentsByContentLike(ctx context.Context, content string) ([]dto.CommentDto, error) {
	var comments []po.Comment

	logger.Info("查询评论数据")
	result := storage.Storage.Db.Model(&po.Comment{}).
		WithContext(ctx).
		Where("LOWER(comment_content) LIKE ?", "%"+strings.ToLower(content)+"%").
		Find(&comments)
	if result.Error != nil {
		msg := fmt.Sprintf("查询评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	logger.Info("查询评论数据成功: %v", result.RowsAffected)

	// 转换为DTO
	var commentDtos []dto.CommentDto
	for _, comment := range comments {
		commentDtos = append(commentDtos, dto.CommentDto{
			CommentId:        comment.CommentId,
			CommenterEmail:   comment.CommenterEmail,
			BlogId:           comment.BlogId,
			OriginPostId:     comment.OriginPostId,
			ReplyToCommentId: comment.ReplyToCommentId,
			Content:          comment.Content,
			CreateTime:       comment.CreateTime,
		})
	}

	return commentDtos, nil
}

// FindCommentsByBlogId 根据博客ID查询评论
// - ctx: 上下文对象
// - blogId: 博客ID
//
// 返回值:
// - []dto.CommentDto: 符合博客ID的评论列表
// - error: 错误信息
func FindCommentsByBlogId(ctx context.Context, blogId string) ([]dto.CommentDto, error) {
	var comments []po.Comment

	logger.Info("根据博客 ID 查询楼主评论数据")
	result := storage.Storage.Db.Model(&po.Comment{}).
		WithContext(ctx).
		Where("blog_id = ? AND (original_poster_id = '' OR original_poster_id IS NULL)", blogId).
		Find(&comments)
	if result.Error != nil {
		msg := fmt.Sprintf("根据博客 ID 查询楼主评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	logger.Info("根据博客 ID 查询楼主评论数据成功: %v", result.RowsAffected)

	// 转换为DTO
	var commentDtos []dto.CommentDto
	for _, comment := range comments {
		commentDtos = append(commentDtos, dto.CommentDto{
			CommentId:        comment.CommentId,
			CommenterEmail:   comment.CommenterEmail,
			BlogId:           comment.BlogId,
			OriginPostId:     comment.OriginPostId,
			ReplyToCommentId: comment.ReplyToCommentId,
			Content:          comment.Content,
			CreateTime:       comment.CreateTime,
		})
	}

	return commentDtos, nil
}

// FindCommentsByOriginPostId 根据楼主评论ID查询评论
// - ctx: 上下文对象
// - originPostId: 楼主评论ID
//
// 返回值:
// - []dto.CommentDto: 符合楼主评论ID的评论列表
// - error: 错误信息
func FindCommentsByOriginPostId(ctx context.Context, originPostId string) ([]dto.CommentDto, error) {
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

	// 转换为DTO
	var commentDtos []dto.CommentDto
	for _, comment := range comments {
		commentDtos = append(commentDtos, dto.CommentDto{
			CommentId:        comment.CommentId,
			CommenterEmail:   comment.CommenterEmail,
			BlogId:           comment.BlogId,
			OriginPostId:     comment.OriginPostId,
			ReplyToCommentId: comment.ReplyToCommentId,
			Content:          comment.Content,
			CreateTime:       comment.CreateTime,
		})
	}

	return commentDtos, nil
}

// FindCommentById 根据评论ID查询单个评论
// - ctx: 上下文对象
// - commentId: 评论ID
//
// 返回值:
// - *dto.CommentDto: 评论数据传输对象
// - error: 错误信息
func FindCommentById(ctx context.Context, commentId string) (*dto.CommentDto, error) {
	var comment po.Comment

	logger.Info("根据评论 ID 查询评论数据")
	result := storage.Storage.Db.Model(&po.Comment{}).
		WithContext(ctx).
		Where("comment_id = ?", commentId).
		First(&comment)
	if result.Error != nil {
		msg := fmt.Sprintf("根据评论 ID 查询评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	logger.Info("根据评论 ID 查询评论数据成功")

	// 转换为DTO
	commentDto := &dto.CommentDto{
		CommentId:        comment.CommentId,
		CommenterEmail:   comment.CommenterEmail,
		BlogId:           comment.BlogId,
		OriginPostId:     comment.OriginPostId,
		ReplyToCommentId: comment.ReplyToCommentId,
		Content:          comment.Content,
		CreateTime:       comment.CreateTime,
	}

	return commentDto, nil
}

// CreateComment 创建新的评论
// - ctx: 上下文对象
// - tx: 数据库事务对象
// - commentDto: 评论数据传输对象
//
// 返回值:
// - *dto.CommentDto: 创建的评论数据传输对象
// - error: 错误信息
func CreateComment(ctx context.Context, tx *gorm.DB, commentDto *dto.CommentDto) (*dto.CommentDto, error) {
	// 生成评论ID
	commentId, err := utils.GenId(fmt.Sprintf("%s_%d", commentDto.CommenterEmail, time.Now().UnixNano()))
	if err != nil {
		msg := fmt.Sprintf("生成评论ID失败: %v", err)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 将DTO转换为PO
	comment := &po.Comment{
		CommentId:        commentId,
		CommenterEmail:   commentDto.CommenterEmail,
		BlogId:           commentDto.BlogId,
		OriginPostId:     commentDto.OriginPostId,
		ReplyToCommentId: commentDto.ReplyToCommentId,
		Content:          commentDto.Content,
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
	}

	logger.Info("添加评论数据")
	// 执行创建操作
	result := tx.Create(comment)
	if result.Error != nil {
		msg := fmt.Sprintf("创建评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	logger.Info("创建评论数据成功: %v", result.RowsAffected)

	// 转换为DTO返回
	resultDto := &dto.CommentDto{
		CommentId:        comment.CommentId,
		CommenterEmail:   comment.CommenterEmail,
		BlogId:           comment.BlogId,
		OriginPostId:     comment.OriginPostId,
		ReplyToCommentId: comment.ReplyToCommentId,
		Content:          comment.Content,
		CreateTime:       comment.CreateTime,
	}

	return resultDto, nil
}

// DeleteCommentById 根据评论ID删除评论
// - ctx: 上下文对象
// - tx: 数据库事务对象
// - id: 评论ID
//
// 返回值:
// - int64: 受影响的行数
// - error: 错误信息
func DeleteCommentById(ctx context.Context, tx *gorm.DB, id string) (int64, error) {
	logger.Info("删除评论数据")
	result := tx.Delete(&po.Comment{CommentId: id})
	if result.Error != nil {
		msg := fmt.Sprintf("删除评论数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	logger.Info("删除评论数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}

// UpdateComment 更新评论
// - ctx: 上下文对象
// - tx: 数据库事务对象
// - commentDto: 评论数据传输对象
//
// 返回值:
// - *dto.CommentDto: 更新后的评论数据传输对象
// - error: 错误信息
func UpdateComment(ctx context.Context, tx *gorm.DB, commentDto *dto.CommentDto) (*dto.CommentDto, error) {
	// 将DTO转换为PO进行更新
	comment := &po.Comment{
		CommentId:        commentDto.CommentId,
		CommenterEmail:   commentDto.CommenterEmail,
		BlogId:           commentDto.BlogId,
		OriginPostId:     commentDto.OriginPostId,
		ReplyToCommentId: commentDto.ReplyToCommentId,
		Content:          commentDto.Content,
		UpdateTime:       time.Now(),
	}

	logger.Info("更新评论数据")
	result := tx.Where("comment_id = ?", comment.CommentId).Updates(comment)
	if result.Error != nil {
		msg := fmt.Sprintf("更新评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	logger.Info("更新评论数据成功: %v", result.RowsAffected)

	// 查询更新后的数据并转换为DTO返回
	var updatedComment po.Comment
	result = tx.Where("comment_id = ?", commentDto.CommentId).First(&updatedComment)
	if result.Error != nil {
		msg := fmt.Sprintf("查询更新后的评论数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 转换为DTO返回
	resultDto := &dto.CommentDto{
		CommentId:        updatedComment.CommentId,
		CommenterEmail:   updatedComment.CommenterEmail,
		BlogId:           updatedComment.BlogId,
		OriginPostId:     updatedComment.OriginPostId,
		ReplyToCommentId: updatedComment.ReplyToCommentId,
		Content:          updatedComment.Content,
		CreateTime:       updatedComment.CreateTime,
	}

	return resultDto, nil
}
