package adminservice

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

func TestGetPresignUrlById(t *testing.T) {
	ctx := context.Background()
	url, err := GetPresignUrlById(ctx, "0ab6f800e0ea3270")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("url = %v\n", url)
	storage.Storage.Close(ctx)
}
