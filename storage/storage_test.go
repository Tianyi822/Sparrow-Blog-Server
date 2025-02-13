package storage

import (
	"context"
	"fmt"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
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
	InitStorage(context.Background())
}

func TestStorage_GetContentFromOss(t *testing.T) {
	ctx := context.Background()
	data, err := Storage.GetContentFromOss(ctx, "articles/C-的智能指针笔记.md")
	if err != nil {
		fmt.Printf("获取内容失败: %v", err)
	}
	fmt.Println(string(data))
}

func TestStorage_ListOssDirFiles(t *testing.T) {
	ctx := context.Background()
	files, err := Storage.ListOssDirFiles(ctx, config.User.ImageOssPath)
	if err != nil {
		fmt.Printf("获取文件列表失败: %v", err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
}

func TestStorage_PreSignUrl(t *testing.T) {
	ctx := context.Background()
	url, err := Storage.PreSignUrl(ctx, "images/1714135012409.webp")
	if err != nil {
		fmt.Printf("获取预签名URL失败: %v", err)
	}

	fmt.Println(url)
}
