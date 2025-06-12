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
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/blevesearch/bleve/v2/search"
)

// 字段名常量，避免硬编码
const (
	FieldID      = "ID"      // 文档ID字段
	FieldTitle   = "Title"   // 标题字段
	FieldContent = "Content" // 内容字段
)

// DefaultSearchFields 默认搜索字段
var DefaultSearchFields = []string{FieldTitle, FieldContent}

// SearchRequest 搜索请求结构
type SearchRequest struct {
	Query     string   `json:"query"`     // 搜索关键词
	Size      int      `json:"size"`      // 返回结果数量，默认10
	From      int      `json:"from"`      // 分页偏移量，默认0
	Fields    []string `json:"fields"`    // 返回字段，默认["Title", "Content"]
	Highlight bool     `json:"highlight"` // 是否启用高亮，默认true
}

// SearchResponse 搜索响应结构
type SearchResponse struct {
	Total  uint64                  `json:"total"`   // 总结果数
	Hits   []*search.DocumentMatch `json:"hits"`    // 搜索结果
	TimeMs float64                 `json:"time_ms"` // 搜索耗时（毫秒）
}

var (
	Index bleve.Index

	loadingOnce sync.Once
)

// Search 执行搜索操作，使用改进的字段特定查询确保中文搜索正常工作
func Search(req SearchRequest) (*SearchResponse, error) {
	// 设置默认值
	if req.Size < 0 {
		req.Size = 10 // 负数时使用默认值
	} else if req.Size == 0 {
		req.Size = 1000 // 0表示返回所有结果，设置一个合理的最大值
	}
	if req.From < 0 {
		req.From = 0
	}
	if len(req.Fields) == 0 {
		req.Fields = DefaultSearchFields
	}

	// 创建字段特定的查询来解决中文搜索问题
	titleQuery := bleve.NewMatchQuery(req.Query)
	titleQuery.SetField(FieldTitle)

	contentQuery := bleve.NewMatchQuery(req.Query)
	contentQuery.SetField(FieldContent)

	// 使用布尔查询组合多个字段查询（Title OR Content）
	boolQuery := bleve.NewBooleanQuery()
	boolQuery.AddShould(titleQuery)
	boolQuery.AddShould(contentQuery)

	// 创建搜索请求
	searchRequest := bleve.NewSearchRequest(boolQuery)
	searchRequest.Size = req.Size
	searchRequest.From = req.From
	searchRequest.Fields = req.Fields

	// 配置高亮
	if req.Highlight {
		highlight := bleve.NewHighlight()
		highlight.AddField(FieldTitle)
		highlight.AddField(FieldContent)
		searchRequest.Highlight = highlight
	}

	// 执行搜索
	searchResult, err := Index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	// 构造响应 - 将时间转换为毫秒
	response := &SearchResponse{
		Total:  searchResult.Total,
		Hits:   searchResult.Hits,
		TimeMs: float64(searchResult.Took) / float64(time.Millisecond),
	}

	return response, nil
}

// LoadingIndex 加载索引
func LoadingIndex(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	loadingOnce.Do(func() {
		// 首先注册中文分词器，确保无论是创建新索引还是加载已存在索引都能正常工作
		if err := registry.RegisterTokenizer("chinese", func(config map[string]any, cache *registry.Cache) (analysis.Tokenizer, error) {
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
				// 索引文章 - 使用IndexedDoc()方法获取正确的文档结构
				if err := index.Index(d.ID, d.IndexedDoc()); err != nil {
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

func CloseIndex() {
	if Index != nil {
		if err := Index.Close(); err != nil {
			logger.Error("关闭索引文件失败: " + err.Error())
		}
	}
}
