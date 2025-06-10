package searchengine

import (
	"context"
	"sparrow_blog_server/internal/repositories/blogrepo"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/filetool"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/searchengine/doc"
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
			logger.Info("创建索引文件")
			chineseMapping, err := mapping.CreateChineseMapping()
			if err != nil {
				logger.Panic("创建中文索引映射失败: " + err.Error())
			}
			index, err := bleve.New(config.SearchEngine.IndexPath, chineseMapping)
			if err != nil {
				logger.Panic("创建索引文件失败: " + err.Error())
			}

			// 生成所有文章的索引
			docs, err := getAllDocs(ctx)
			if err != nil {
				logger.Panic("生成所有文章的索引失败: " + err.Error())
			}
			for _, d := range docs {
				err := d.GetContent(ctx)
				if err != nil {
					logger.Error("获取文章内容失败: " + err.Error())
					continue
				}
				// 索引文章
				if err := index.Index(d.ID, d); err != nil {
					logger.Error("索引文章失败: " + err.Error())
				} else {
					logger.Info("索引文章成功: " + d.Title)
				}
			}

			Index = index
		}
	})

	return nil
}

// getAllDocs 获取所有文章
func getAllDocs(ctx context.Context) ([]doc.Doc, error) {
	blogDtos, err := blogrepo.FindAllBlogs(ctx, false)
	if err != nil {
		return nil, err
	}

	docs := make([]doc.Doc, len(blogDtos))
	for i, blogDto := range blogDtos {
		docs[i] = doc.Doc{
			ID:    blogDto.BlogId,
			Title: blogDto.BlogTitle,
		}
	}

	return docs, nil
}

func CloseIndex() error {
	if Index != nil {
		if err := Index.Close(); err != nil {
			return err
		}
	}

	return nil
}
