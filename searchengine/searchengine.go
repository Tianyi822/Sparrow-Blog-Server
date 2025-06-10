package searchengine

import (
	"context"
	"github.com/blevesearch/bleve/v2"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/filetool"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/searchengine/mapping"
	"sync"
)

var (
	Index bleve.Index

	loadingOnce sync.Once
)

// LoadingIndex 加载索引
func LoadingIndex(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	loadingOnce.Do(func() {
		if filetool.IsExist(config.SearchEngine.IndexPath) {
			logger.Info("加载本地索引文件")
			index, err := bleve.Open(config.SearchEngine.IndexPath)
			if err != nil {
				logger.Panic("加载本地索引文件失败: " + err.Error())
			}
			Index = index
		} else {
			// TODO: 这里应该需要扫描所有文章，重建索引，但是先空着
			logger.Info("创建索引文件")
			chineseMapping, err := mapping.CreateChineseMapping()
			if err != nil {
				logger.Panic("创建中文索引映射失败: " + err.Error())
			}
			index, err := bleve.New(config.SearchEngine.IndexPath, chineseMapping)
			if err != nil {
				logger.Panic("创建索引文件失败: " + err.Error())
			}
			Index = index
		}
	})

	return nil
}

func CloseIndex() error {
	if Index != nil {
		if err := Index.Close(); err != nil {
			return err
		}
	}

	return nil
}
