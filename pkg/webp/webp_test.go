package webp

import (
	"context"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/storage"
	"testing"
	"time"
)

func init() {
	// 加载配置文件
	config.LoadConfig("../../resources/config/test/pkg-config.yaml")
	// 初始化 Logger 组件
	err := logger.InitLogger()
	if err != nil {
		return
	}
	// 初始化数据库组件
	storage.InitStorage()
	// 初始化 WebP 转换器
	InitConverter()
}

func TestConverter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	files, err := storage.Storage.ListOssDirFiles(ctx, config.UserConfig.ImageOssPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatal("没有图片文件")
	}

	Converter.AddBatchTasks(ctx, files)

	time.Sleep(6 * time.Minute)
}
