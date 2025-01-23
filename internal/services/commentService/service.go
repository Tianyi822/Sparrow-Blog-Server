package commentService

import (
	"context"
	"h2blog/internal/model/vo"
	"h2blog/internal/repository/commentRepo"
)

// GetCommentsByBlogId 根据博客ID获取评论
func GetCommentsByBlogId(ctx context.Context, blogId string) (*vo.CommentsVo, error) {
	// 获取评论
	comments, err := commentRepo.FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		return nil, err
	}

	// 用于保存评论
	commentsVo := &vo.CommentsVo{}
	for _, comment := range comments {
		// 获取子评论
		subComments, err := commentRepo.FindCommentsByParentId(ctx, comment.CommentId)
		if err != nil {
			return nil, err
		}
		// 将子评论转为 Vo，并保存
		for _, subComment := range subComments {
			commentsVo.SubComments = append(commentsVo.SubComments, vo.CommentVo{
				CommentId:  subComment.CommentId,
				Content:    subComment.Content,
				UserName:   subComment.UserName,
				UserEmail:  subComment.UserEmail,
				UserUrl:    subComment.UserUrl,
				CreateTime: subComment.CreateTime,
			})
		}
	}

	return commentsVo, nil
}
