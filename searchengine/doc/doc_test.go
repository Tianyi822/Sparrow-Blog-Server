package doc

import (
	"context"
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

func TestDoc_GetContent(t *testing.T) {
	doc := Doc{
		ID:    "e84f1230f7358390",
		Title: "新博客-new",
	}

	content, err := doc.GetContent(context.Background())
	if err != nil {
		t.Error(err)
	}
	t.Log(string(content))
}
