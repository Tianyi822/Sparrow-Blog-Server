package ossstore

import (
	"fmt"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
)

const (
	MarkDown = "markdown"
	Webp     = "webp"
	JPG      = "jpg"
	JPEG     = "jpeg"
	PNG      = "png"
)

const (
	Get = "GET"
	Put = "PUT"
)

// GenOssSavePath 用于生成博客保存路径
func GenOssSavePath(name string, fileType string) string {
	switch fileType {
	case MarkDown:
		return fmt.Sprintf("%s%s.md", config.Oss.BlogOssPath, name)
	case Webp:
		return fmt.Sprintf("%s%s.webp", config.Oss.ImageOssPath, name)
	case JPG:
		return fmt.Sprintf("%s%s.jpg", config.Oss.ImageOssPath, name)
	case JPEG:
		return fmt.Sprintf("%s%s.jpeg", config.Oss.ImageOssPath, name)
	case PNG:
		return fmt.Sprintf("%s%s.png", config.Oss.ImageOssPath, name)
	default:
		logger.Error("不存在该文件类型")
		return ""
	}
}
