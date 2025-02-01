package webp

import (
	"context"
	"h2blog/internal/model/dto"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/storage"
	"h2blog/storage/oss"
	"strings"
	"testing"
	"time"
)

func init() {
	// 加载配置文件
	config.LoadConfig("../../resources/config/test/pkg-config.yaml")
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	storage.InitStorage(context.Background())
	// 初始化 WebP 转换器
	InitConverter(context.Background())
}

func TestConverter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	files, err := storage.Storage.ListOssDirFiles(ctx, config.UserConfig.ImageOssPath)
	if err != nil {
		t.Fatal(err)
	}

	imgDtos := make([]dto.ImgDto, 0, len(files))
	for _, file := range files {
		imgInfo := strings.Split(file, "/")[1]
		imgName := strings.Split(imgInfo, ".")[0]
		imgType := strings.Split(imgInfo, ".")[1]
		imgDto := dto.ImgDto{
			ImgName: imgName,
		}
		switch imgType {
		case "jpg":
			imgDto.ImgType = oss.JPG
		case "jpeg":
			imgDto.ImgType = oss.JPEG
		case "png":
			imgDto.ImgType = oss.PNG
		default:
			continue
		}
		imgDtos = append(imgDtos, imgDto)
	}

	Converter.AddBatchTasks(ctx, imgDtos)

	time.Sleep(6 * time.Minute)
}
