package oss

import (
	"fmt"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
)

const (
	MarkDown = "markdown"
	HTML     = "html"
	Webp     = "webp"
	JPG      = "jpg"
	JPEG     = "jpeg"
	PNG      = "png"
)

// GenOssSavePath 用于生成博客保存路径
func GenOssSavePath(name string, fileType string) string {
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
