package adminservices

import (
	"context"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/repositories/commentrepo"
	"sparrow_blog_server/internal/services/webservice"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/utils"
	"sparrow_blog_server/storage"
	"testing"
	"time"
)

func init() {
	// 加载配置文件
	config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	_ = storage.InitStorage(context.Background())
}

// 模拟数据
func setupTestData() (*dto.CommentDto, string) {
	// 生成测试用的博客ID
	blogId, _ := utils.GenId(fmt.Sprintf("test_blog_%d", time.Now().UnixNano()))

	// 创建测试评论DTO
	commentDto := &dto.CommentDto{
		CommenterEmail:   "test@example.com",
		BlogId:           blogId,
		OriginPostId:     "", // 楼主评论
		ReplyToCommentId: "", // 不回复任何评论
		Content:          "这是一条测试评论",
	}

	return commentDto, blogId
}

// TestAddComment 测试添加评论
func TestAddComment(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据
	commentDto, _ := setupTestData()

	// 调用添加评论方法
	commentVo, err := webservice.AddComment(ctx, commentDto)
	if err != nil {
		t.Fatalf("添加评论失败: %v", err)
	}

	// 验证返回结果
	if commentVo == nil {
		t.Fatal("返回的评论对象为空")
	}

	if commentVo.CommentId == "" {
		t.Error("评论ID不能为空")
	}

	if commentVo.CommenterEmail != commentDto.CommenterEmail {
		t.Errorf("评论者邮箱不匹配: 期望 %s, 实际 %s", commentDto.CommenterEmail, commentVo.CommenterEmail)
	}

	if commentVo.Content != commentDto.Content {
		t.Errorf("评论内容不匹配: 期望 %s, 实际 %s", commentDto.Content, commentVo.Content)
	}

	// 清理测试数据
	cleanupTx := storage.Storage.Db.WithContext(ctx).Begin()
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, commentVo.CommentId)
	cleanupTx.Commit()

	t.Logf("添加评论测试通过: ID=%s", commentVo.CommentId)
}

// TestUpdateComment 测试更新评论
func TestUpdateComment(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据 - 先添加一条评论
	commentDto, _ := setupTestData()
	commentVo, err := webservice.AddComment(ctx, commentDto)
	if err != nil {
		t.Fatalf("准备测试数据失败: %v", err)
	}

	// 准备更新数据
	updateDto := &dto.CommentDto{
		Content: "这是更新后的评论内容",
	}

	// 调用更新评论方法
	updatedVo, err := UpdateComment(ctx, commentVo.CommentId, updateDto)
	if err != nil {
		t.Fatalf("更新评论失败: %v", err)
	}

	// 验证更新结果
	if updatedVo == nil {
		t.Fatal("返回的更新评论对象为空")
	}

	if updatedVo.Content != updateDto.Content {
		t.Errorf("评论内容更新失败: 期望 %s, 实际 %s", updateDto.Content, updatedVo.Content)
	}

	if updatedVo.CommentId != commentVo.CommentId {
		t.Errorf("评论ID不匹配: 期望 %s, 实际 %s", commentVo.CommentId, updatedVo.CommentId)
	}

	// 清理测试数据
	cleanupTx := storage.Storage.Db.WithContext(ctx).Begin()
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, commentVo.CommentId)
	cleanupTx.Commit()

	t.Logf("更新评论测试通过: ID=%s", updatedVo.CommentId)
}

// TestDeleteComment 测试删除评论
func TestDeleteComment(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据 - 先添加一条评论
	commentDto, _ := setupTestData()
	commentVo, err := webservice.AddComment(ctx, commentDto)
	if err != nil {
		t.Fatalf("准备测试数据失败: %v", err)
	}

	// 调用删除评论方法
	err = DeleteComment(ctx, commentVo.CommentId)
	if err != nil {
		t.Fatalf("删除评论失败: %v", err)
	}

	// 验证删除结果 - 尝试查找已删除的评论应该失败
	_, err = commentrepo.FindCommentById(ctx, commentVo.CommentId)
	if err == nil {
		t.Error("评论应该已被删除，但仍然可以找到")
	}

	t.Logf("删除评论测试通过: ID=%s", commentVo.CommentId)
}

// TestGetCommentsByBlogId 测试根据博客ID获取评论
func TestGetCommentsByBlogId(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据 - 添加楼主评论和子评论
	commentDto, blogId := setupTestData()

	// 添加楼主评论
	mainCommentVo, err := webservice.AddComment(ctx, commentDto)
	if err != nil {
		t.Fatalf("添加楼主评论失败: %v", err)
	}

	// 添加子评论
	subCommentDto := &dto.CommentDto{
		CommenterEmail:   "sub@example.com",
		BlogId:           blogId,
		OriginPostId:     "",                      // 这会在服务层根据ReplyToCommentId自动设置
		ReplyToCommentId: mainCommentVo.CommentId, // 回复楼主评论
		Content:          "这是一条子评论",
	}

	subCommentVo, err := webservice.AddComment(ctx, subCommentDto)
	if err != nil {
		t.Fatalf("添加子评论失败: %v", err)
	}

	// 调用获取评论方法
	comments, err := webservice.GetCommentsByBlogId(ctx, blogId)
	if err != nil {
		t.Fatalf("获取评论失败: %v", err)
	}

	// 验证结果
	if len(comments) == 0 {
		t.Fatal("应该至少有一条评论")
	}

	// 检查楼主评论
	found := false
	for _, comment := range comments {
		if comment.CommentId == mainCommentVo.CommentId {
			found = true

			// 检查是否有子评论
			if len(comment.SubComments) == 0 {
				t.Error("楼主评论应该有子评论")
			} else {
				// 验证子评论
				subFound := false
				for _, subComment := range comment.SubComments {
					if subComment.CommentId == subCommentVo.CommentId {
						subFound = true
						break
					}
				}
				if !subFound {
					t.Error("未找到预期的子评论")
				}
			}
			break
		}
	}

	if !found {
		t.Error("未找到预期的楼主评论")
	}

	// 清理测试数据
	cleanupTx := storage.Storage.Db.WithContext(ctx).Begin()
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, mainCommentVo.CommentId)
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, subCommentVo.CommentId)
	cleanupTx.Commit()

	t.Logf("获取评论测试通过: 博客ID=%s, 评论数=%d", blogId, len(comments))
}

// TestAddCommentWithInvalidData 测试添加评论时的错误处理
func TestAddCommentWithInvalidData(t *testing.T) {
	ctx := context.Background()

	// 测试空的评论内容
	commentDto := &dto.CommentDto{
		CommenterEmail: "test@example.com",
		BlogId:         "test_blog_id",
		Content:        "", // 空内容
	}

	// 这个测试主要验证服务能够处理空内容的情况
	// 在实际应用中，可能需要在服务层添加验证逻辑
	commentVo, err := webservice.AddComment(ctx, commentDto)
	if err != nil {
		t.Logf("添加空内容评论失败（符合预期）: %v", err)
	} else {
		// 如果成功创建，则清理数据
		if commentVo != nil {
			cleanupTx := storage.Storage.Db.WithContext(ctx).Begin()
			_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, commentVo.CommentId)
			cleanupTx.Commit()
		}
		t.Log("添加空内容评论成功")
	}
}

// TestUpdateNonExistentComment 测试更新不存在的评论
func TestUpdateNonExistentComment(t *testing.T) {
	ctx := context.Background()

	// 使用不存在的评论ID
	nonExistentId := "non_existent_comment_id"
	updateDto := &dto.CommentDto{
		Content: "尝试更新不存在的评论",
	}

	// 调用更新方法，应该失败
	_, err := UpdateComment(ctx, nonExistentId, updateDto)
	if err == nil {
		t.Error("更新不存在的评论应该失败，但却成功了")
	} else {
		t.Logf("更新不存在的评论失败（符合预期）: %v", err)
	}
}

// TestDeleteNonExistentComment 测试删除不存在的评论
func TestDeleteNonExistentComment(t *testing.T) {
	ctx := context.Background()

	// 使用不存在的评论ID
	nonExistentId := "non_existent_comment_id"

	// 调用删除方法，应该失败
	err := DeleteComment(ctx, nonExistentId)
	if err == nil {
		t.Error("删除不存在的评论应该失败，但却成功了")
	} else {
		t.Logf("删除不存在的评论失败（符合预期）: %v", err)
	}
}

// TestAddReplyComment 测试回复评论功能
func TestAddReplyComment(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据 - 先添加一条楼主评论
	commentDto, blogId := setupTestData()
	mainCommentVo, err := webservice.AddComment(ctx, commentDto)
	if err != nil {
		t.Fatalf("添加楼主评论失败: %v", err)
	}

	// 创建回复评论
	replyCommentDto := &dto.CommentDto{
		CommenterEmail:   "reply@example.com",
		BlogId:           blogId,
		OriginPostId:     "",                      // 这会在服务层自动设置
		ReplyToCommentId: mainCommentVo.CommentId, // 回复楼主评论
		Content:          "这是一条回复评论",
	}

	// 添加回复评论
	replyCommentVo, err := webservice.AddComment(ctx, replyCommentDto)
	if err != nil {
		t.Fatalf("添加回复评论失败: %v", err)
	}

	// 验证回复评论的字段
	if replyCommentVo.ReplyToCommentId != mainCommentVo.CommentId {
		t.Errorf("回复评论ID不正确: 期望 %s, 实际 %s", mainCommentVo.CommentId, replyCommentVo.ReplyToCommentId)
	}

	if replyCommentVo.OriginPostId != mainCommentVo.CommentId {
		t.Errorf("楼主评论ID不正确: 期望 %s, 实际 %s", mainCommentVo.CommentId, replyCommentVo.OriginPostId)
	}

	// 创建回复回复的评论（二级回复）
	replyToReplyDto := &dto.CommentDto{
		CommenterEmail:   "reply2@example.com",
		BlogId:           blogId,
		OriginPostId:     "",                       // 这会在服务层自动设置
		ReplyToCommentId: replyCommentVo.CommentId, // 回复刚才的回复评论
		Content:          "这是一条回复回复的评论",
	}

	// 添加二级回复评论
	replyToReplyVo, err := webservice.AddComment(ctx, replyToReplyDto)
	if err != nil {
		t.Fatalf("添加二级回复评论失败: %v", err)
	}

	// 验证二级回复评论的字段
	if replyToReplyVo.ReplyToCommentId != replyCommentVo.CommentId {
		t.Errorf("二级回复评论ID不正确: 期望 %s, 实际 %s", replyCommentVo.CommentId, replyToReplyVo.ReplyToCommentId)
	}

	// 二级回复的OriginPostId应该和一级回复的OriginPostId相同（都指向楼主评论）
	if replyToReplyVo.OriginPostId != mainCommentVo.CommentId {
		t.Errorf("二级回复的楼主评论ID不正确: 期望 %s, 实际 %s", mainCommentVo.CommentId, replyToReplyVo.OriginPostId)
	}

	// 清理测试数据
	cleanupTx := storage.Storage.Db.WithContext(ctx).Begin()
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, mainCommentVo.CommentId)
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, replyCommentVo.CommentId)
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, replyToReplyVo.CommentId)
	cleanupTx.Commit()

	t.Logf("回复评论测试通过: 楼主评论ID=%s, 回复评论ID=%s, 二级回复ID=%s",
		mainCommentVo.CommentId, replyCommentVo.CommentId, replyToReplyVo.CommentId)
}

// TestDeleteCommentWithSubComments 测试删除评论及其子评论功能
func TestDeleteCommentWithSubComments(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据 - 先添加一条楼主评论
	commentDto, blogId := setupTestData()
	mainCommentVo, err := webservice.AddComment(ctx, commentDto)
	if err != nil {
		t.Fatalf("添加楼主评论失败: %v", err)
	}

	// 创建回复评论
	replyCommentDto := &dto.CommentDto{
		CommenterEmail:   "reply@example.com",
		BlogId:           blogId,
		OriginPostId:     "",                      // 这会在服务层自动设置
		ReplyToCommentId: mainCommentVo.CommentId, // 回复楼主评论
		Content:          "这是一条回复评论",
	}

	// 添加回复评论
	replyCommentVo, err := webservice.AddComment(ctx, replyCommentDto)
	if err != nil {
		t.Fatalf("添加回复评论失败: %v", err)
	}

	// 创建二级回复评论
	replyToReplyDto := &dto.CommentDto{
		CommenterEmail:   "reply2@example.com",
		BlogId:           blogId,
		OriginPostId:     "",                       // 这会在服务层自动设置
		ReplyToCommentId: replyCommentVo.CommentId, // 回复刚才的回复评论
		Content:          "这是一条二级回复评论",
	}

	// 添加二级回复评论
	replyToReplyVo, err := webservice.AddComment(ctx, replyToReplyDto)
	if err != nil {
		t.Fatalf("添加二级回复评论失败: %v", err)
	}

	// 调用删除评论及子评论方法
	err = DeleteCommentWithSubComments(ctx, mainCommentVo.CommentId)
	if err != nil {
		t.Fatalf("删除评论及子评论失败: %v", err)
	}

	// 验证删除结果 - 尝试查找已删除的评论应该都失败
	_, err = commentrepo.FindCommentById(ctx, mainCommentVo.CommentId)
	if err == nil {
		t.Error("主评论应该已被删除，但仍然可以找到")
	}

	_, err = commentrepo.FindCommentById(ctx, replyCommentVo.CommentId)
	if err == nil {
		t.Error("回复评论应该已被删除，但仍然可以找到")
	}

	_, err = commentrepo.FindCommentById(ctx, replyToReplyVo.CommentId)
	if err == nil {
		t.Error("二级回复评论应该已被删除，但仍然可以找到")
	}

	t.Logf("删除评论及子评论测试通过: 主评论ID=%s", mainCommentVo.CommentId)
}

// TestGetAllComments 测试获取所有评论功能
func TestGetAllComments(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据 - 添加几条评论
	commentDto1, _ := setupTestData()
	commentVo1, err := webservice.AddComment(ctx, commentDto1)
	if err != nil {
		t.Fatalf("添加第一条评论失败: %v", err)
	}

	commentDto2, _ := setupTestData()
	commentDto2.CommenterEmail = "test2@example.com"
	commentDto2.Content = "这是第二条测试评论"
	commentVo2, err := webservice.AddComment(ctx, commentDto2)
	if err != nil {
		t.Fatalf("添加第二条评论失败: %v", err)
	}

	// 调用获取所有评论方法
	comments, err := GetAllComments(ctx)
	if err != nil {
		t.Fatalf("获取所有评论失败: %v", err)
	}

	// 验证结果
	if len(comments) < 2 {
		t.Errorf("应该至少有2条评论，实际获取到%d条", len(comments))
	}

	// 检查是否包含我们添加的评论
	found1 := false
	found2 := false
	for _, comment := range comments {
		if comment.CommentId == commentVo1.CommentId {
			found1 = true
		}
		if comment.CommentId == commentVo2.CommentId {
			found2 = true
		}
	}

	if !found1 {
		t.Error("未找到第一条测试评论")
	}
	if !found2 {
		t.Error("未找到第二条测试评论")
	}

	// 清理测试数据
	cleanupTx := storage.Storage.Db.WithContext(ctx).Begin()
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, commentVo1.CommentId)
	_, _ = commentrepo.DeleteCommentById(ctx, cleanupTx, commentVo2.CommentId)
	cleanupTx.Commit()

	t.Logf("获取所有评论测试通过: 总评论数=%d", len(comments))
}
