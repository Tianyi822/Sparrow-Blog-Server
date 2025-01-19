package oss

import (
	"fmt"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
)

type FileType string

func (ft FileType) String() string {
	return string(ft)
}

const (
	MarkDown FileType = "markdown"
	HTML     FileType = "html"
	Webp     FileType = "webp"
	JPG      FileType = "jpg"
	JPEG     FileType = "jpeg"
	PNG      FileType = "png"
)

// GenOssSavePath 用于生成博客保存路径
func GenOssSavePath(name string, fileType FileType) string {
	switch fileType {
	case MarkDown:
		return fmt.Sprintf("%s%s.md", config.UserConfig.BlogOssPath, name)
	case HTML:
		return fmt.Sprintf("%s%s.html", config.UserConfig.BlogOssPath, name)
	case Webp:
		return fmt.Sprintf("%s%s.webp", config.UserConfig.ImageOssPath, name)
	case JPG:
		return fmt.Sprintf("%s%s.jpg", config.UserConfig.ImageOssPath, name)
	case JPEG:
		return fmt.Sprintf("%s%s.jpeg", config.UserConfig.ImageOssPath, name)
	case PNG:
		return fmt.Sprintf("%s%s.png", config.UserConfig.ImageOssPath, name)
	default:
		logger.Error("不存在该文件类型")
		return ""
	}
}
