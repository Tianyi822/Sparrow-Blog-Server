package email

import (
	"context"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"testing"
)

func init() {
	config.LoadConfig()
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

func TestSendVerificationCodeEmail(t *testing.T) {
	ctx := context.Background()
	err := SendVerificationCodeBySys(ctx)
	if err != nil {
		t.Fatalf("SendVerificationCodeBySys failed: %v", err)
	}

	err = SendVerificationCodeBySys(ctx)
	if err != nil {
		t.Fatalf("SendVerificationCodeBySys failed: %v", err)
	}
}
