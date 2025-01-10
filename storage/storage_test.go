package storage

import (
	"context"
	"fmt"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"testing"
)

func init() {
	// 加载配置文件
	config.LoadConfig("../resources/config/test/storage-config.yaml")
	// 初始化 Logger 组件
	err := logger.InitLogger()
	if err != nil {
		return
	}
	// 初始化数据库组件
	InitStorage()
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
	files, err := Storage.ListOssDirFiles(ctx, config.UserConfig.ImageOssPath)
	if err != nil {
		fmt.Printf("获取文件列表失败: %v", err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
}
