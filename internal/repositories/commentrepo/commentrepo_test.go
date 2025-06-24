package commentrepo

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
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
