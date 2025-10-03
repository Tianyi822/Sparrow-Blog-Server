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



func TestSendCommentNotificationEmail(t *testing.T) {
	ctx := context.Background()

	// 创建测试评论数据
	comment := CommentData{
		CommenterEmail: "test@example.com",
		BlogTitle:      "Go语言学习笔记",
		Content:        "这篇文章写得非常好，对Go语言的并发编程讲解得很清楚，特别是goroutine和channel的部分。期待更多这样的技术分享！",
		CreateTime:     "2024-06-24 10:30:00",
	}

	err := SendCommentNotificationBySys(ctx, comment)
	if err != nil {
		t.Fatalf("SendCommentNotificationBySys failed: %v", err)
	}
}

func TestSendReplyNotificationEmail(t *testing.T) {
	ctx := context.Background()

	// 创建测试回复数据
	reply := ReplyData{
		ReplierEmail:    "replier@example.com",
		BlogTitle:       "Go语言学习笔记",
		OriginalContent: "这篇文章写得非常好，对Go语言的并发编程讲解得很清楚。",
		ReplyContent:    "我也有同样的感受！特别是关于select语句的使用，让我对Go的并发模型有了更深的理解。博主能再分享一些实际项目中的应用案例吗？",
		CreateTime:      "2024-06-24 11:15:00",
	}

	err := SendReplyNotificationBySys(ctx, reply)
	if err != nil {
		t.Fatalf("SendReplyNotificationBySys failed: %v", err)
	}
}

func TestSendCommentOrReplyNotificationEmail(t *testing.T) {
	ctx := context.Background()

	// 测试评论通知（replyToCommentId 为空）
	t.Run("发送评论通知", func(t *testing.T) {
		err := SendCommentOrReplyNotification(
			ctx,
			"newcommenter@example.com",
			"Docker容器化实践指南",
			"这篇Docker教程非常实用，按照步骤操作很容易上手。希望能看到更多关于Kubernetes的内容！",
			"2024-06-24 14:20:00",
			"", // 空字符串表示这是评论而不是回复
			"",
			"",
		)
		if err != nil {
			t.Errorf("发送评论通知失败: %v", err)
		} else {
			t.Log("评论通知发送成功")
		}
	})

	// 测试回复通知（replyToCommentId 不为空）
	t.Run("发送回复通知", func(t *testing.T) {
		err := SendCommentOrReplyNotification(
			ctx,
			"replier@example.com",
			"Docker容器化实践指南",
			"关于Kubernetes的内容确实很有必要，我推荐先从基本的Pod和Service概念开始学习。",
			"2024-06-24 15:30:00",
			"comment_id_123", // 非空表示这是回复
			"这篇Docker教程非常实用，希望能看到更多关于Kubernetes的内容！",
			"newcommenter@example.com",
		)
		if err != nil {
			t.Errorf("发送回复通知失败: %v", err)
		} else {
			t.Log("回复通知发送成功")
		}
	})

	// 测试自己回复自己的情况（不应该发送通知）
	t.Run("自己回复自己不发送通知", func(t *testing.T) {
		err := SendCommentOrReplyNotification(
			ctx,
			"same@example.com",
			"Docker容器化实践指南",
			"补充一下我之前的评论，还有一些细节需要注意。",
			"2024-06-24 16:00:00",
			"comment_id_456",
			"我之前发表的评论内容",
			"same@example.com", // 相同邮箱，不应该发送通知
		)
		if err != nil {
			t.Errorf("处理自己回复自己时出错: %v", err)
		} else {
			t.Log("自己回复自己，正确地未发送通知")
		}
	})
}
