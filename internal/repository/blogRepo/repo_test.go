package blogRepo

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

func TestAddBlog(t *testing.T) {
	blogDto := dto.BlogDto{
		BlogTitle:    "测试博客",
		BlogBrief:    "测试博客",
		CategoryId:   "category00001",
		BlogWordsNum: 100,
		BlogState:    true,
		BlogIsTop:    false,
	}
	err := AddBlog(context.Background(), &blogDto)

	if err != nil {
		t.Error(err)
	}

	t.Log(blogDto.BlogId)
}

func TestFindBlogsInPage(t *testing.T) {
	page, err := FindAllBlogs(context.Background(), false)
	if err != nil {
		t.Error(err)
	}

	for _, v := range page {
		t.Log(v)
	}
}

func TestDeleteBlogById(t *testing.T) {
	err := DeleteBlogById(context.Background(), "blog00002")

	if err != nil {
		t.Error(err)
	}
}

func TestChangeBlogStateById(t *testing.T) {
	err := ChangeBlogStateById(context.Background(), "blog00003")
	if err != nil {
		t.Error(err)
	}
}

func TestSetTopById(t *testing.T) {
	err := SetTopById(context.Background(), "blog00003")
	if err != nil {
		t.Error(err)
	}
}
