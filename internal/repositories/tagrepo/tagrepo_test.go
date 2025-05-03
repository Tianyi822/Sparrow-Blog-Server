package tagrepo

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
	"testing"
)

func init() {
	// 加载配置文件
	_ = config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	_ = storage.InitStorage(context.Background())
}

func TestFindTagsByBlogId(t *testing.T) {
	ctx := context.Background()

	tags, err := FindTagsByBlogId(ctx, "blog00011")

	if err != nil {
		t.Errorf("FindTagsByBlogId() error = %v", err)
		return
	}

	for _, tag := range tags {
		t.Logf("tag: %v", tag)
	}
}

func TestAddTags(t *testing.T) {
	ctx := context.Background()

	tx := storage.Storage.Db.WithContext(ctx).Begin()
	tags := []dto.TagDto{
		{
			TagName: "tag00001",
		},
		{
			TagName: "tag00002",
		},
	}

	newTags, err := AddTags(tx, tags)
	tx.Commit()

	if err != nil {
		t.Errorf("AddTags() error = %v", err)
	}

	for _, tag := range newTags {
		t.Logf("tag: %v", tag)
	}
}

func TestDeleteTags(t *testing.T) {
	ctx := context.Background()

	tx := storage.Storage.Db.WithContext(ctx).Begin()
	tags := []dto.TagDto{
		{
			TagId:   "e7c039f52925c96f",
			TagName: "tag00001",
		},
		{
			TagId:   "b30fd680a54768d7",
			TagName: "tag00002",
		},
	}
	err := DeleteTags(tx, tags)
	tx.Commit()

	if err != nil {
		t.Errorf("DeleteTags() error = %v", err)
	}
}

func TestAddBlogTagAssociation(t *testing.T) {
	ctx := context.Background()

	tx := storage.Storage.Db.WithContext(ctx).Begin()

	err := AddBlogTagAssociation(tx, "blog00003", []dto.TagDto{
		{
			TagId: "e7c039f52925c96f",
		},
		{
			TagId: "b30fd680a54768d7",
		},
	})

	tx.Commit()

	if err != nil {
		t.Errorf("AddBlogTagAssociation() error = %v", err)
	}
}

func TestGetAllTags(t *testing.T) {
	ctx := context.Background()

	tags, err := FindAllTags(ctx)

	if err != nil {
		t.Errorf("FindAllTags() error = %v", err)
	}

	for _, tag := range tags {
		t.Logf("tag: %v", tag)
	}
}
