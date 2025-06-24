package commentrepo

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestAddComment(t *testing.T) {
	tests := []struct {
		name       string
		commentDto dto.CommentDto
		wantErr    bool
	}{
		{
			name: "正常添加评论",
			commentDto: dto.CommentDto{
				CommenterEmail:   "test@example.com",
				BlogId:           "test_blog_1",
				OriginPostId:     "",
				ReplyToCommentId: "",
				Content:          "Test comment content",
			},
			wantErr: false,
		},
		{
			name: "添加子评论",
			commentDto: dto.CommentDto{
				CommenterEmail:   "test2@example.com",
				BlogId:           "test_blog_1",
				OriginPostId:     "parent_1",
				ReplyToCommentId: "reply_to_1",
				Content:          "Test sub comment",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 开启事务
			tx := storage.Storage.Db.Begin()
			defer tx.Rollback()

			// 执行测试
			resultDto, err := CreateComment(tx, &tt.commentDto)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resultDto)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultDto)
				assert.NotEmpty(t, resultDto.CommentId)
				assert.Equal(t, tt.commentDto.CommenterEmail, resultDto.CommenterEmail)
				assert.Equal(t, tt.commentDto.Content, resultDto.Content)
				assert.Equal(t, tt.commentDto.BlogId, resultDto.BlogId)
				assert.Equal(t, tt.commentDto.OriginPostId, resultDto.OriginPostId)
				assert.Equal(t, tt.commentDto.ReplyToCommentId, resultDto.ReplyToCommentId)

				// 验证数据是否正确保存
				var saved po.Comment
				err = tx.Where("comment_id = ?", resultDto.CommentId).First(&saved).Error
				assert.NoError(t, err)
				assert.Equal(t, resultDto.Content, saved.Content)
			}
		})
	}
}

func TestDeleteCommentById(t *testing.T) {
	// 创建测试用的CommentDto
	commentDto := &dto.CommentDto{
		CommenterEmail:   "test@example.com",
		BlogId:           "test_blog_1",
		OriginPostId:     "",
		ReplyToCommentId: "",
		Content:          "Test comment for deletion",
	}
	// 开启事务用于创建测试数据
	setupTx := storage.Storage.Db.Begin()
	defer setupTx.Rollback()

	resultDto, err := CreateComment(setupTx, commentDto)
	assert.NoError(t, err)
	setupTx.Commit()

	tests := []struct {
		name     string
		id       string
		wantRows int64
		wantErr  bool
	}{
		{
			name:     "正常删除评论",
			id:       resultDto.CommentId, // 使用实际创建的ID
			wantRows: 1,
			wantErr:  false,
		},
		{
			name:     "删除不存在的评论",
			id:       "non_existent",
			wantRows: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data
			tx := storage.Storage.Db.Begin()
			defer tx.Rollback()

			// Execute test
			rows, err := DeleteCommentById(tx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, rows)

				if tt.wantRows > 0 {
					// Verify deletion
					var count int64
					tx.Model(&po.Comment{}).Where("comment_id = ?", tt.id).Count(&count)
					assert.Equal(t, int64(0), count)
				}
			}
		})
	}
}

// TestDeleteCommentsByBlogId 测试根据博客ID删除所有评论
func TestDeleteCommentsByBlogId(t *testing.T) {
	ctx := context.Background()
	blogId := "test_blog_del"

	// 创建测试数据 - 多个评论
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

	// 创建测试评论
	var createdComments []*dto.CommentDto
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	for _, comment := range comments {
		created, err := CreateComment(tx, comment)
		if err != nil {
			tx.Rollback()
			t.Fatalf("创建测试评论失败: %v", err)
		}
		createdComments = append(createdComments, created)
	}
	tx.Commit()

	// 验证评论已创建
	foundComments, err := FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		t.Fatalf("查询评论失败: %v", err)
	}
	if len(foundComments) != 3 {
		t.Fatalf("期望创建3条评论，实际创建%d条", len(foundComments))
	}

	// 测试删除所有评论
	deleteTx := storage.Storage.Db.WithContext(ctx).Begin()
	rowsAffected, err := DeleteCommentsByBlogId(deleteTx, blogId)
	if err != nil {
		deleteTx.Rollback()
		t.Fatalf("删除评论失败: %v", err)
	}
	deleteTx.Commit()

	// 验证删除结果
	if rowsAffected != 3 {
		t.Errorf("期望删除3条评论，实际删除%d条", rowsAffected)
	}

	// 验证评论已被删除
	remainingComments, err := FindCommentsByBlogId(ctx, blogId)
	if err != nil {
		t.Fatalf("查询剩余评论失败: %v", err)
	}
	if len(remainingComments) != 0 {
		t.Errorf("期望删除后无剩余评论，实际剩余%d条", len(remainingComments))
	}

	t.Logf("成功删除博客%s的所有评论", blogId)
}
