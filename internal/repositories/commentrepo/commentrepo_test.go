package commentrepo

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"strings"
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

func TestAddComment(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		comment  po.Comment
		wantRows int64
		wantErr  bool
	}{
		{
			name: "正常添加评论",
			comment: po.Comment{
				CommentId:  "test_comment_1",
				Content:    "Test comment content",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantRows: 1,
			wantErr:  false,
		},
		{
			name: "重复评论ID",
			comment: po.Comment{
				CommentId:  "test_comment_1",
				Content:    "Duplicate comment",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantRows: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 开启事务
			tx := storage.Storage.Db.Begin()
			defer tx.Rollback()

			// 执行测试
			rows, err := AddComment(ctx, &tt.comment)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantRows, rows)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, rows)

				// 验证数据是否正确保存
				var saved po.Comment
				err = tx.Where("comment_id = ?", tt.comment.CommentId).First(&saved).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.comment.Content, saved.Content)
			}
		})
	}
}

func TestDeleteCommentById(t *testing.T) {
	ctx := context.Background()

	// Create test data
	comment := &po.Comment{
		CommentId:  "test_comment_1",
		Content:    "Test comment for deletion",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	_, err := AddComment(ctx, comment)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		id       string
		wantRows int64
		wantErr  bool
	}{
		{
			name:     "正常删除评论",
			id:       "test_comment_1",
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
			rows, err := DeleteCommentById(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, rows)

				// Verify deletion
				var count int64
				tx.Model(&po.Comment{}).Where("comment_id = ?", tt.id).Count(&count)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}

func TestFindCommentByContentLike(t *testing.T) {
	ctx := context.Background()

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
		_, err := AddComment(ctx, &comment)
		if err != nil {
			t.Fatal(err)
		}
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
