package commentrepo

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"strings"
	"testing"
	"time"

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
	ctx := context.Background()

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
			resultDto, err := CreateComment(ctx, tx, &tt.commentDto)

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
	ctx := context.Background()

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

	resultDto, err := CreateComment(ctx, setupTx, commentDto)
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
			rows, err := DeleteCommentById(ctx, tx, tt.id)

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

func TestFindCommentByContentLike(t *testing.T) {
	ctx := context.Background()

	// 清理可能存在的测试数据
	cleanupTx := storage.Storage.Db.Begin()
	cleanupTx.Exec("DELETE FROM COMMENT WHERE commenter_email = 'test@example.com'")
	cleanupTx.Commit()

	// Create test data
	testComments := []po.Comment{
		{
			CommentId:  "test_comment_1",
			Content:    "Test comment content",
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
		{
			CommentId:  "test_comment_2",
			Content:    "Another test content",
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
	}

	for _, comment := range testComments {
		// 创建测试用的CommentDto
		commentDto := &dto.CommentDto{
			CommenterEmail:   "test@example.com",
			BlogId:           "test_blog_1",
			OriginPostId:     "",
			ReplyToCommentId: "",
			Content:          comment.Content,
		}
		// 开启事务用于创建测试数据
		setupTx := storage.Storage.Db.Begin()
		_, err := CreateComment(ctx, setupTx, commentDto)
		if err != nil {
			setupTx.Rollback()
			t.Fatal(err)
		}
		setupTx.Commit()
	}

	tests := []struct {
		name      string
		content   string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Find existing comments",
			content:   "test",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "Find specific comment",
			content:   "Another",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "Find non-existent comment",
			content:   "nonexistent",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := FindCommentsByContentLike(ctx, tt.content)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(comments))

				if tt.wantCount > 0 {
					for _, comment := range comments {
						assert.Contains(t, strings.ToLower(comment.Content), strings.ToLower(tt.content))
					}
				}
			}
		})
	}
}
