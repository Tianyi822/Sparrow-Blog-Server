package commentservice

import (
	"context"
	"h2blog_server/internal/model/vo"
	"h2blog_server/internal/repositories/commentrepo"
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
	comments, err := commentrepo.FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		return nil, err
	}

	// 保存所有楼主评论
	var commentVos []vo.CommentVo

	// 遍历所有楼主评论
	for _, comment := range comments {
		// 用于保存评论
		commentVo := vo.CommentVo{}
		// 获取楼层子评论
		subComments, err := commentrepo.FindCommentsByOriginPostId(ctx, comment.CommentId)
		if err != nil {
			return nil, err
		}
		// 将子评论转为 Vo，并保存
		for _, subComment := range subComments {
			commentVo.SubComments = append(commentVo.SubComments, vo.CommentVo{
				CommentId:  subComment.CommentId,
				Content:    subComment.Content,
				UserName:   subComment.CommenterName,
				UserEmail:  subComment.CommenterEmail,
				UserUrl:    subComment.CommenterUrl,
				CreateTime: subComment.CreateTime,
			})
		}
		// 添加到楼主评论集合
		commentVos = append(commentVos, commentVo)
	}

	return commentVos, nil
}
