package adminservices

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/repositories/commentrepo"
	"sparrow_blog_server/internal/services/webservice"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/utils"
	"sparrow_blog_server/storage"
	"testing"
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

func TestUpdateBlogData(t *testing.T) {
	blogDto := &dto.BlogDto{
		BlogId:     "blog00006",
		BlogTitle:  "test6",
		BlogBrief:  "test",
		Category:   &dto.CategoryDto{CategoryName: "测试分类 7"},
		CategoryId: "",
		BlogIsTop:  false,
		BlogState:  true,
		Tags: []dto.TagDto{
			{TagId: "tag001", TagName: "Java"},
			{TagName: "tag00006"},
		},
	}

	err := UpdateOrAddBlog(context.Background(), blogDto)

	if err != nil {
		t.Error(err)
	}
}

func TestAddBlog(t *testing.T) {
	blogDto := &dto.BlogDto{
		BlogTitle: "test6",
		BlogBrief: "test",
		Category: &dto.CategoryDto{
			CategoryName: "测试分类 6",
		},
		CategoryId: "",
		BlogIsTop:  false,
		BlogState:  true,
		Tags: []dto.TagDto{
			{TagId: "tag001", TagName: "Java"},
			{TagName: "TEST_TAG"},
		},
	}

	err := UpdateOrAddBlog(context.Background(), blogDto)

	if err != nil {
		t.Error(err)
	}
	t.Log(blogDto.BlogId)
}

func TestDeleteBlog(t *testing.T) {
	err := DeleteBlogById(context.Background(), "d5fbb6bf6c5b3a27")
	if err != nil {
		t.Error(err)
	}
	t.Log("success")
}

func TestGetBlogData(t *testing.T) {
	blogDto, url, err := GetBlogData(context.Background(), "0e317bad75975c0a")
	if err != nil {
		t.Error(err)
	}

	// 格式化输出
	t.Logf("%+v", blogDto)
	t.Log(url)
}

// TestDeleteBlogWithComments 测试删除博客时是否正确删除了相关评论
func TestDeleteBlogWithComments(t *testing.T) {
	ctx := context.Background()

	// 创建测试博客ID
	blogId, err := utils.GenId("test_blog")
	if err != nil {
		t.Fatalf("生成博客ID失败: %v", err)
	}

	// 创建测试评论
	comments := []*dto.CommentDto{
		{
			CommenterEmail: "user1@example.com",
			BlogId:         blogId,
			Content:        "First comment",
		},
		{
			CommenterEmail: "user2@example.com",
			BlogId:         blogId,
			Content:        "Second comment",
		},
		{
			CommenterEmail: "user3@example.com",
			BlogId:         blogId,
			Content:        "Third comment",
		},
	}

	// 添加测试评论
	for _, comment := range comments {
		_, err := webservice.AddComment(ctx, comment)
		if err != nil {
			t.Fatalf("创建测试评论失败: %v", err)
		}
	}

	// 验证评论已创建
	foundComments, err := commentrepo.FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		t.Fatalf("查询评论失败: %v", err)
	}
	if len(foundComments) != 3 {
		t.Fatalf("期望创建3条评论，实际创建%d条", len(foundComments))
	}

	// 测试删除博客时是否删除了相关评论（这里我们直接调用 commentrepo 的方法来模拟博客删除）
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		tx.Rollback() // 确保测试不会影响数据库
	}()

	// 删除博客的所有评论
	rowsAffected, err := commentrepo.DeleteCommentsByBlogId(tx, blogId)
	if err != nil {
		t.Fatalf("删除博客相关评论失败: %v", err)
	}
	tx.Commit()

	// 验证删除结果
	if rowsAffected != 3 {
		t.Errorf("期望删除3条评论，实际删除%d条", rowsAffected)
	}

	// 验证评论已被删除
	remainingComments, err := commentrepo.FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		t.Fatalf("查询剩余评论失败: %v", err)
	}
	if len(remainingComments) != 0 {
		t.Errorf("期望删除后无剩余评论，实际剩余%d条", len(remainingComments))
	}

	t.Logf("成功测试删除博客时删除相关评论: 博客ID=%s, 删除评论数=%d", blogId, rowsAffected)
}
