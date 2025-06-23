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

func TestSendFriendLinkNotificationEmail(t *testing.T) {
	ctx := context.Background()

	// 创建测试友链数据
	friendLink := FriendLinkData{
		Name:        "测试博客",
		URL:         "https://example.com",
		AvatarURL:   "https://example.com/avatar.jpg",
		Description: "这是一个非常棒的测试博客，包含了很多有趣的技术文章和教程。",
	}

	err := SendFriendLinkNotificationBySys(ctx, friendLink)
	if err != nil {
		t.Fatalf("SendFriendLinkNotificationBySys failed: %v", err)
	}
}
