package blogRepo

import (
	"context"
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

func TestFindBlogsInPage(t *testing.T) {
	page, err := FindBlogsInPage(context.Background(), 1, 10)
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
