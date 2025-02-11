package oss

import (
	"fmt"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
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
		return fmt.Sprintf("%s%s.md", config.User.BlogOssPath, name)
	case HTML:
		return fmt.Sprintf("%s%s.html", config.User.BlogOssPath, name)
	case Webp:
		return fmt.Sprintf("%s%s.webp", config.User.ImageOssPath, name)
	case JPG:
		return fmt.Sprintf("%s%s.jpg", config.User.ImageOssPath, name)
	case JPEG:
		return fmt.Sprintf("%s%s.jpeg", config.User.ImageOssPath, name)
	case PNG:
		return fmt.Sprintf("%s%s.png", config.User.ImageOssPath, name)
	default:
		logger.Error("不存在该文件类型")
		return ""
	}
}
