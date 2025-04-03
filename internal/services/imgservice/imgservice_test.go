package imgservice

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/webp"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
	"strings"
	"testing"
	"time"
)

func init() {
	// 加载配置文件
	_ = config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	_ = storage.InitStorage(context.Background())
	// 初始化转换器
	_ = webp.InitConverter(context.Background())
}

func TestConvertAndAddImg(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	files, err := storage.Storage.ListOssDirFiles(ctx, config.Oss.ImageOssPath)
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
			imgDto.ImgType = ossstore.JPG
		case "jpeg":
			imgDto.ImgType = ossstore.JPEG
		case "png":
			imgDto.ImgType = ossstore.PNG
		default:
			continue
		}
		imgDtos = append(imgDtos, imgDto)
	}

	imgsVo, err := ConvertAndAddImg(ctx, imgDtos)
	if err != nil {
		t.Fatal(err)
	}
	for _, imgVo := range imgsVo.Success {
		t.Logf("imgVo Successed: %#v\n", imgVo)
	}
	for _, imgVo := range imgsVo.Failure {
		t.Logf("imgVo Failed: %#v\n", imgVo)
	}
	t.Logf("success len = %v\n", len(imgsVo.Success))
	t.Logf("fail len = %v\n", len(imgsVo.Failure))
}

func TestGetPresignUrlById(t *testing.T) {
	ctx := context.Background()
	url, err := GetPresignUrlById(ctx, "0ab6f800e0ea3270")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("url = %v\n", url)
	storage.Storage.Close(ctx)
}
