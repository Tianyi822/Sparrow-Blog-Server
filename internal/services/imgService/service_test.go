package imgService

import (
	"context"
	"h2blog/internal/model/dto"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/pkg/utils"
	"h2blog/pkg/webp"
	"h2blog/storage"
	"strings"
	"testing"
	"time"
)

func init() {
	// 加载配置文件
	config.LoadConfig("../../../resources/config/test/service-config.yaml")
	// 初始化 Logger 组件
	err := logger.InitLogger()
	if err != nil {
		return
	}
	// 初始化数据库组件
	storage.InitStorage()
	// 初始化转换器
	webp.InitConverter()
}

func TestConvertAndAddImg(t *testing.T) {
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
			imgDto.ImgType = utils.JPG
		case "jpeg":
			imgDto.ImgType = utils.JPEG
		case "png":
			imgDto.ImgType = utils.PNG
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
	for _, imgVo := range imgsVo.Fail {
		t.Logf("imgVo Failed: %#v\n", imgVo)
	}
	t.Logf("success len = %v\n", len(imgsVo.Success))
	t.Logf("fail len = %v\n", len(imgsVo.Fail))
}
