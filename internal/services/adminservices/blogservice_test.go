package adminservices

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
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
