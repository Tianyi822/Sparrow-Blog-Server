package blogService

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/webp"
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
	// 初始化转换器
	_ = webp.InitConverter(context.Background())
}

func TestGetBlogsInPage(t *testing.T) {
	blogDtos, err := GetBlogsInPage(context.Background(), 1, 10)

	if err != nil {
		t.Error(err)
	}

	for _, blogDto := range blogDtos {
		t.Log(blogDto)
	}
}

func TestUpdateBlogData(t *testing.T) {
	blogDto := &dto.BlogDto{
		BId:        "blog00006",
		Title:      "test6",
		Brief:      "test",
		Category:   dto.CategoryDto{CName: "测试分类 6"},
		CategoryId: "",
		IsTop:      false,
		State:      true,
		Tags: []dto.TagDto{
			{TId: "e7c039f52925c96f", TName: "tag00001"},
			{TName: "tag00006"},
		},
	}

	err := UpdateBlogData(context.Background(), blogDto)

	if err != nil {
		t.Error(err)
	}
}
