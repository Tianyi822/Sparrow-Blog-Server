package searchengine

import (
	"context"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/filetool"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/searchengine/mapping"
	"sparrow_blog_server/searchengine/tokenizer"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/registry"
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
		// 首先注册中文分词器，确保无论是创建新索引还是加载已存在索引都能正常工作
		if err := registry.RegisterTokenizer("chinese", func(config map[string]interface{}, cache *registry.Cache) (analysis.Tokenizer, error) {
			return tokenizer.NewChineseTokenizer(), nil
		}); err != nil {
			logger.Panic("注册中文分词器失败: " + err.Error())
		}

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
