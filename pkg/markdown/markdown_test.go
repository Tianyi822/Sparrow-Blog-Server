package markdown

import (
	"context"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/storage"
	"os"
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
	storage.InitStorage(context.Background())
	// 初始化 Markdown 渲染器
	InitRenderer(context.Background())
}

func TestCustomRenderer(t *testing.T) {
	ctx := context.Background()
	data, err := storage.Storage.GetContentFromOss(ctx, "articles/C-的智能指针笔记.md")
	if err != nil {
		t.Errorf("获取内容失败: %v", err)
	}

	result, err := Parse(data)
	if err != nil {
		t.Errorf("解析Markdown失败: %v", err)
	}

	// 写入结果到文件
	if err := os.WriteFile("../../temp/rendered_markdown.html", result, 0644); err != nil {
		t.Errorf("写入文件失败: %v", err)
	}

	t.Logf("渲染结果已写入: %s", "../../temp/rendered_markdown.html")
}
