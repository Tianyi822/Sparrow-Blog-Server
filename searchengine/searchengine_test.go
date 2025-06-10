package searchengine

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

func TestIndex(t *testing.T) {
	// 初始化搜索引擎组件
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	stats := Index.Stats()
	t.Logf("索引统计信息: %v", stats)
}
