package searchengine

import (
	"context"
	"fmt"
	"path/filepath"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/repositories/blogrepo"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/filetool"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/searchengine/doc"
	"sparrow_blog_server/searchengine/mapping"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	blevemapping "github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search"
)

// 字段名常量，避免硬编码
const (
	FieldID      = "ID"      // 文档 ID 字段
	FieldImgId   = "ImgId"   // 图片 ID 字段
	FieldTitle   = "Title"   // 标题字段
	FieldContent = "Content" // 内容字段
)

// DefaultSearchFields 默认搜索字段
var DefaultSearchFields = []string{FieldID, FieldImgId, FieldTitle, FieldContent}

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
		// 检查索引路径配置
		if config.SearchEngine.IndexPath == "" {
			logger.Panic("搜索引擎索引路径未配置")
		}

		// 记录索引路径信息
		logger.Info("索引路径: " + config.SearchEngine.IndexPath)

		if filetool.IsExist(config.SearchEngine.IndexPath) {
			logger.Info("加载本地索引文件")

			// 检查索引目录权限
			indexDir := filepath.Dir(config.SearchEngine.IndexPath)
			if err := filetool.CheckDirPermissions(indexDir); err != nil {
				logger.Panic("索引目录权限检查失败: " + err.Error())
			}

			index, err := bleve.Open(config.SearchEngine.IndexPath)
			if err != nil {
				logger.Panic("加载本地索引文件失败: " + err.Error())
			}
			Index = index
		} else {
			logger.Info("创建索引文件")

			// 确保索引文件的目录存在并检查权限
			indexDir := filepath.Dir(config.SearchEngine.IndexPath)
			logger.Info("准备创建索引目录: " + indexDir)

			if err := filetool.EnsureDir(indexDir); err != nil {
				logger.Panic("创建索引目录失败: " + err.Error())
			}

			logger.Info("索引目录创建成功，开始创建索引映射")

			// 创建使用内置 unicode 分词器的映射
			unicodeMapping, err := mapping.CreateChineseMapping()
			if err != nil {
				logger.Panic("创建 Unicode 索引映射失败: " + err.Error())
			}

			logger.Info("开始创建索引文件: " + config.SearchEngine.IndexPath)

			index, err := createIndexSafely(config.SearchEngine.IndexPath, unicodeMapping)
			if err != nil {
				logger.Panic("创建索引文件失败: " + err.Error())
			}

			logger.Info("索引文件创建成功，开始建立文档索引")

			// 生成所有文章的索引
			docs, err := getAllDocs(ctx)
			if err != nil {
				logger.Panic("生成所有文章的索引失败: " + err.Error())
			}

			logger.Info("开始为 " + fmt.Sprintf("%d", len(docs)) + " 篇文章建立索引")

			successCount := 0
			for i, d := range docs {
				err := d.GetContent(ctx)
				if err != nil {
					logger.Error("获取文章内容失败: " + err.Error())
					continue
				}
				// 索引文章 - 使用IndexedDoc()方法获取正确的文档结构
				if err := index.Index(d.ID, d.IndexedDoc()); err != nil {
					logger.Error("索引文章失败: " + err.Error())
				} else {
					successCount++
					if (i+1)%10 == 0 || i == len(docs)-1 {
						logger.Info("索引进度: " + fmt.Sprintf("%d/%d", i+1, len(docs)))
					}
				}
			}

			logger.Info("索引建立完成，成功索引文章数: " + fmt.Sprintf("%d", successCount))
			Index = index
		}
	})

	return nil
}

// createIndexSafely 安全地创建索引（带重试机制）
func createIndexSafely(indexPath string, indexMapping blevemapping.IndexMapping) (bleve.Index, error) {
	var lastErr error

	// 尝试3次创建索引
	for attempt := 1; attempt <= 3; attempt++ {
		logger.Info("尝试创建索引，第 " + fmt.Sprintf("%d", attempt) + " 次")

		index, err := bleve.New(indexPath, indexMapping)
		if err == nil {
			logger.Info("索引创建成功")
			return index, nil
		}

		lastErr = err
		logger.Error("索引创建失败，第 " + fmt.Sprintf("%d", attempt) + " 次: " + err.Error())

		// 如果不是最后一次尝试，清理可能损坏的文件
		if attempt < 3 {
			logger.Info("清理可能损坏的索引文件")
			if filetool.IsExist(indexPath) {
				if removeErr := filetool.ForceRemove(indexPath); removeErr != nil {
					logger.Error("清理索引文件失败: " + removeErr.Error())
				}
			}

			// 短暂等待后重试
			logger.Info("等待 1 秒后重试")
			time.Sleep(time.Second)
		}
	}

	return nil, fmt.Errorf("创建索引失败，已尝试3次: %w", lastErr)
}

// getAllDocs 获取所有文章
func getAllDocs(ctx context.Context) ([]doc.Doc, error) {
	blogDtos, err := blogrepo.FindAllBlogs(ctx, true)
	if err != nil {
		return nil, err
	}

	docs := make([]doc.Doc, len(blogDtos))
	for i, blogDto := range blogDtos {
		docs[i] = doc.Doc{
			ID:    blogDto.BlogId,
			ImgId: blogDto.BlogImageId,
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

// AddIndex 将单个博客文档添加到搜索索引中
// 参数:
//   - ctx: 上下文，用于取消操作和超时控制
//   - blogDto: 博客数据传输对象，包含博客的基本信息
//
// 返回值:
//   - error: 如果添加失败则返回错误，成功则返回 nil
func AddIndex(ctx context.Context, blogDto *dto.BlogDto) error {
	// 检查索引是否已初始化
	if Index == nil {
		return fmt.Errorf("搜索索引未初始化")
	}

	// 检查必要的参数
	if blogDto == nil {
		return fmt.Errorf("博客数据不能为空")
	}

	if blogDto.BlogId == "" {
		return fmt.Errorf("博客ID不能为空")
	}

	// 创建 Doc 对象
	d := doc.Doc{
		ID:    blogDto.BlogId,
		ImgId: blogDto.BlogImageId,
		Title: blogDto.BlogTitle,
	}

	// 获取博客内容
	if err := d.GetContent(ctx); err != nil {
		logger.Error("获取博客内容失败 ID = " + d.ID + ": " + err.Error())
		return fmt.Errorf("获取博客内容失败: %w", err)
	}

	// 将文档添加到索引中
	if err := Index.Index(d.ID, d.IndexedDoc()); err != nil {
		logger.Error("索引博客失败 ID = " + d.ID + ": " + err.Error())
		return fmt.Errorf("索引博客失败: %w", err)
	}

	logger.Info("成功添加博客到索引: " + d.Title + " (ID: " + d.ID + ")")
	return nil
}

// UpdateIndex 更新搜索索引中的博客文档
// 参数:
//   - ctx: 上下文，用于取消操作和超时控制
//   - blogDto: 博客数据传输对象，包含更新后的博客信息
//
// 返回值:
//   - error: 如果更新失败则返回错误，成功则返回 nil
func UpdateIndex(ctx context.Context, blogDto *dto.BlogDto) error {
	// 检查索引是否已初始化
	if Index == nil {
		return fmt.Errorf("搜索索引未初始化")
	}

	// 检查必要的参数
	if blogDto == nil {
		return fmt.Errorf("博客数据不能为空")
	}

	if blogDto.BlogId == "" {
		return fmt.Errorf("博客ID不能为空")
	}

	// 创建 Doc 对象
	d := doc.Doc{
		ID:    blogDto.BlogId,
		ImgId: blogDto.BlogImageId,
		Title: blogDto.BlogTitle,
	}

	// 获取博客内容
	if err := d.GetContent(ctx); err != nil {
		logger.Error("获取博客内容失败 ID = " + d.ID + ": " + err.Error())
		return fmt.Errorf("获取博客内容失败: %w", err)
	}

	// 更新索引中的文档（Bleve的Index方法会自动覆盖已存在的文档）
	if err := Index.Index(d.ID, d.IndexedDoc()); err != nil {
		logger.Error("更新博客索引失败 ID = " + d.ID + ": " + err.Error())
		return fmt.Errorf("更新博客索引失败: %w", err)
	}

	logger.Info("成功更新博客索引: " + d.Title + " (ID: " + d.ID + ")")
	return nil
}

// DeleteIndex 从搜索索引中删除博客文档
// 参数:
//   - blogId: 要删除的博客ID
//
// 返回值:
//   - error: 如果删除失败则返回错误，成功则返回 nil
func DeleteIndex(blogId string) error {
	// 检查索引是否已初始化
	if Index == nil {
		return fmt.Errorf("搜索索引未初始化")
	}

	// 检查必要的参数
	if blogId == "" {
		return fmt.Errorf("博客ID不能为空")
	}

	// 从索引中删除文档
	if err := Index.Delete(blogId); err != nil {
		logger.Error("删除博客索引失败 ID = " + blogId + ": " + err.Error())
		return fmt.Errorf("删除博客索引失败: %w", err)
	}

	logger.Info("成功删除博客索引 ID: " + blogId)
	return nil
}

// RebuildIndex 重建搜索索引。
// 该函数会删除现有索引文件，重新创建索引并重新索引所有文档。
// 适用于索引损坏、映射更新或需要完全重建索引的场景。
//
// 参数：
//   - ctx: 上下文，用于取消操作和超时控制
//
// 返回值：
//   - error: 如果重建失败则返回错误，成功则返回 nil
func RebuildIndex(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger.Info("开始重建搜索索引")

	// 1. 关闭现有索引
	if Index != nil {
		if err := Index.Close(); err != nil {
			logger.Error("关闭现有索引失败: " + err.Error())
		}
		Index = nil
	}

	// 2. 删除现有索引文件
	if filetool.IsExist(config.SearchEngine.IndexPath) {
		logger.Info("删除现有索引文件")
		if err := filetool.ForceRemove(config.SearchEngine.IndexPath); err != nil {
			return err
		}
	}

	// 3. 创建新的索引映射
	logger.Info("创建新的索引映射")
	unicodeMapping, err := mapping.CreateChineseMapping()
	if err != nil {
		return err
	}

	// 4. 创建新索引
	logger.Info("创建新索引文件")

	// 确保索引文件的目录存在并检查权限
	indexDir := filepath.Dir(config.SearchEngine.IndexPath)
	logger.Info("准备创建索引目录: " + indexDir)

	if err := filetool.EnsureDir(indexDir); err != nil {
		return fmt.Errorf("创建索引目录失败: %w", err)
	}

	logger.Info("索引目录权限验证通过，开始创建新索引")

	newIndex, err := bleve.New(config.SearchEngine.IndexPath, unicodeMapping)
	if err != nil {
		return fmt.Errorf("创建新索引失败: %w", err)
	}

	// 5. 获取所有文档
	logger.Info("获取所有文档数据")
	docs, err := getAllDocs(ctx)
	if err != nil {
		if closeErr := newIndex.Close(); closeErr != nil {
			logger.Error("关闭新索引失败: " + closeErr.Error())
		}
		return err
	}

	// 6. 重新索引所有文档
	logger.Info("开始重新索引所有文档")
	successCount := 0
	errorCount := 0

	for i, d := range docs {
		select {
		case <-ctx.Done():
			if closeErr := newIndex.Close(); closeErr != nil {
				logger.Error("关闭新索引失败: " + closeErr.Error())
			}
			return ctx.Err()
		default:
		}

		// 获取文档内容
		if err := d.GetContent(ctx); err != nil {
			logger.Error("获取文档内容失败 ID = " + d.ID + ": " + err.Error())
			errorCount++
			continue
		}

		// 索引文档
		if err := newIndex.Index(d.ID, d.IndexedDoc()); err != nil {
			logger.Error("索引文档失败 ID = " + d.ID + ": " + err.Error())
			errorCount++
		} else {
			successCount++
			if (i+1)%100 == 0 || i == len(docs)-1 {
				logger.Info("索引进度: " + fmt.Sprintf("%d/%d", i+1, len(docs)))
			}
		}
	}

	// 7. 更新全局索引引用
	Index = newIndex

	logger.Info("重建索引完成")
	logger.Info("成功索引文档数: " + fmt.Sprintf("%d", successCount))
	if errorCount > 0 {
		logger.Warn("索引失败文档数: " + fmt.Sprintf("%d", errorCount))
	}

	return nil
}
