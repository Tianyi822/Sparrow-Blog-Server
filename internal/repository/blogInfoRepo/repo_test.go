package blogInfoRepo

import (
	"context"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
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
	storage.InitStorage(context.Background())
}

func TestCreateBlogInfo(t *testing.T) {
	tests := []struct {
		name     string
		blogInfo *po.BlogInfo
		wantRows int64
		wantErr  bool
	}{
		{
			name: "successful create",
			blogInfo: &po.BlogInfo{
				BlogId:     "test-blog-1",
				Title:      "Test Blog",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantRows: 1,
			wantErr:  false,
		},
		{
			name: "duplicate blog id",
			blogInfo: &po.BlogInfo{
				BlogId: "test-blog-1", // Same ID as above
				Title:  "Duplicate Blog",
			},
			wantRows: 0,
			wantErr:  true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := CreateBlogInfo(ctx, tt.blogInfo)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBlogInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && rows != tt.wantRows {
				t.Errorf("CreateBlogInfo() rows = %v, want %v", rows, tt.wantRows)
			}
		})
	}

	// Cleanup
	storage.Storage.Db.Where("blog_id LIKE ?", "test-blog-%").Delete(&po.BlogInfo{})
}

func TestFindBlogById(t *testing.T) {
	// Create test blog first
	ctx := context.Background()
	testBlog := &po.BlogInfo{
		BlogId:     "test-find-blog-1",
		Title:      "Test Find Blog",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	_, err := CreateBlogInfo(ctx, testBlog)
	if err != nil {
		t.Fatalf("Failed to create test blog: %v", err)
	}

	tests := []struct {
		name    string
		blogId  string
		want    *po.BlogInfo
		wantErr bool
	}{
		{
			name:   "existing blog",
			blogId: "test-find-blog-1",
			want: &po.BlogInfo{
				BlogId: "test-find-blog-1",
				Title:  "Test Find Blog",
			},
			wantErr: false,
		},
		{
			name:    "non-existing blog",
			blogId:  "non-existing-id",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindBlogById(ctx, tt.blogId)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindBlogById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.BlogId != tt.want.BlogId {
					t.Errorf("FindBlogById() got = %v, want %v", got.BlogId, tt.want.BlogId)
				}
				if got.Title != tt.want.Title {
					t.Errorf("FindBlogById() got = %v, want %v", got.Title, tt.want.Title)
				}
			}
		})
	}

	// Cleanup
	storage.Storage.Db.Where("blog_id LIKE ?", "test-find-blog-%").Delete(&po.BlogInfo{})
}

func TestUpdateById(t *testing.T) {
	// Create test blog first
	ctx := context.Background()
	testBlog := &po.BlogInfo{
		BlogId:     "test-update-blog",
		Title:      "Test Update Blog",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	_, err := CreateBlogInfo(ctx, testBlog)
	if err != nil {
		t.Fatalf("Failed to create test blog: %v", err)
	}

	tests := []struct {
		name     string
		blogInfo *po.BlogInfo
		wantRows int64
		wantErr  bool
	}{
		{
			name: "successful update",
			blogInfo: &po.BlogInfo{
				BlogId:     "test-update-blog",
				Title:      "Updated Test Blog",
				UpdateTime: time.Now(),
			},
			wantRows: 1,
			wantErr:  false,
		},
		{
			name: "non-existing blog",
			blogInfo: &po.BlogInfo{
				BlogId: "non-existing-id",
				Title:  "Non-existing Blog",
			},
			wantRows: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := UpdateById(ctx, tt.blogInfo)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if rows != tt.wantRows {
					t.Errorf("UpdateById() rows = %v, want %v", rows, tt.wantRows)
				}

				// Verify the update
				updated, _ := FindBlogById(ctx, tt.blogInfo.BlogId)
				if updated.Title != tt.blogInfo.Title {
					t.Errorf("UpdateById() got title = %v, want %v", updated.Title, tt.blogInfo.Title)
				}
			}
		})
	}

	// Cleanup
	storage.Storage.Db.Where("blog_id LIKE ?", "test-update-blog-%").Delete(&po.BlogInfo{})
}

func TestDeleteById(t *testing.T) {
	// Create test blog first
	ctx := context.Background()
	testBlog := &po.BlogInfo{
		BlogId:     "test-delete-blog",
		Title:      "Test Delete Blog",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	_, err := CreateBlogInfo(ctx, testBlog)
	if err != nil {
		t.Fatalf("Failed to create test blog: %v", err)
	}

	tests := []struct {
		name     string
		blogId   string
		wantRows int64
		wantErr  bool
	}{
		{
			name:     "successful delete",
			blogId:   "test-delete-blog",
			wantRows: 1,
			wantErr:  false,
		},
		{
			name:     "non-existing blog",
			blogId:   "non-existing-id",
			wantRows: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := DeleteById(ctx, tt.blogId)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if rows != tt.wantRows {
					t.Errorf("DeleteById() rows = %v, want %v", rows, tt.wantRows)
				}

				// Verify the deletion
				_, err := FindBlogById(ctx, tt.blogId)
				if err == nil {
					t.Errorf("DeleteById() blog still exists after deletion")
				}
			}
		})
	}

	// Cleanup
	storage.Storage.Db.Where("blog_id LIKE ?", "test-delete-blog-%").Delete(&po.BlogInfo{})
}
