package webservice

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

func TestGetHomeData(t *testing.T) {
	data, err := GetHomeData(context.Background())
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%#v", data)
	}
}

// TestGetLatestComments 测试获取最新评论功能
func TestGetLatestComments(t *testing.T) {
	ctx := context.Background()

	// 调用服务方法获取最新评论
	comments, err := GetLatestComments(ctx)
	if err != nil {
		t.Fatalf("获取最新评论失败: %v", err)
	}

	// 验证结果
	if len(comments) > 5 {
		t.Errorf("期望最多返回5条评论，实际返回%d条", len(comments))
	}

	// 验证评论按时间倒序排列
	for i := 1; i < len(comments); i++ {
		if comments[i-1].CreateTime.Before(comments[i].CreateTime) {
			t.Errorf("评论排序错误：第%d条评论的时间晚于第%d条", i+1, i)
		}
	}

	t.Logf("成功获取最新%d条评论", len(comments))
	for i, comment := range comments {
		t.Logf("第%d条评论: ID=%s, Email=%s, BlogTitle=%s, Content=%s, Time=%s",
			i+1, comment.CommentId, comment.CommenterEmail, comment.BlogTitle,
			comment.Content[:min(20, len(comment.Content))],
			comment.CreateTime.Format("2006-01-02 15:04:05"))
	}
}

// min 辅助函数，返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
