package storage

import (
	"context"
	"fmt"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage/ossstore"
	"testing"
	"time"
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
	_ = InitStorage(context.Background())
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
	files, err := Storage.ListOssDirFiles(ctx, config.Oss.ImageOssPath)
	if err != nil {
		fmt.Printf("获取文件列表失败: %v", err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
}

func TestStorage_PreSignUrl(t *testing.T) {
	ctx := context.Background()
	url, err := Storage.GenPreSignUrl(ctx, "images/1714135012409.webp", ossstore.Webp, ossstore.Get, 10*time.Minute)
	if err != nil {
		fmt.Printf("获取预签名URL失败: %v", err)
	}

	fmt.Println(url)
}

func TestStorage_PreSignUrl_Put(t *testing.T) {
	ctx := context.Background()
	url, err := Storage.GenPreSignUrl(ctx, "images/1714135012409.webp", ossstore.Webp, ossstore.Put, 1*time.Minute)
	if err != nil {
		fmt.Printf("获取预签名URL失败: %v", err)
	}

	fmt.Println(url)
}
