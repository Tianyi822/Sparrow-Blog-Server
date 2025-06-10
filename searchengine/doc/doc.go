package doc

import (
	"context"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"sparrow_blog_server/storage/ossstore"
)

type Doc struct {
	ID      string // 文档 ID
	Title   string // 文档标题
	Content []byte // 文档内容
}

// GetContent 从OSS中获取文档的Markdown内容
// 参数:
//   - ctx context.Context: 上下文对象，用于传递请求范围的 deadline、取消信号、认证信息等
//
// 返回值:
//   - error: 错误对象，如果获取内容时发生错误则返回
func (d *Doc) GetContent(ctx context.Context) error {
	logger.Info("从OSS中获取文档内容")
	content, err := storage.Storage.GetContentFromOss(ctx, ossstore.GenOssSavePath(d.Title, ossstore.MarkDown))
	if err != nil {
		return err
	}
	d.Content = content

	return nil
}
