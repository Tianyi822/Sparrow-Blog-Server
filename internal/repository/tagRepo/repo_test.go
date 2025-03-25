package tagRepo

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

	tags := []dto.TagDto{
		{
			TName: "tag00001",
		},
		{
			TName: "tag00002",
		},
	}

	_, err := AddTags(ctx, tags)

	if err != nil {
		t.Errorf("AddTags() error = %v", err)
	}
}

func TestAddBlogTagAssociation(t *testing.T) {
	ctx := context.Background()

	err := AddBlogTagAssociation(ctx, "blog00003", []dto.TagDto{
		{
			TId: "e7c039f52925c96f",
		},
		{
			TId: "b30fd680a54768d7",
		},
	})

	if err != nil {
		t.Errorf("AddBlogTagAssociation() error = %v", err)
	}
}
