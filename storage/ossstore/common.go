package ossstore

import (
	"fmt"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
)

const (
	MarkDown = "markdown"
	Webp     = "webp"
)

const (
	Get = "GET"
	Put = "PUT"
)

const (
	MarkdownHeader = "text/markdown"
	WebpHeader     = "image/webp"
)

// GenOssSavePath 用于生成博客保存路径
func GenOssSavePath(name string, fileType string) string {
	switch fileType {
	case MarkDown:
		return fmt.Sprintf("%s%s.md", config.Oss.BlogOssPath, name)
	case Webp:
		return fmt.Sprintf("%s%s.webp", config.Oss.ImageOssPath, name)
	default:
		logger.Error("不存在该文件类型")
		return ""
	}
}
