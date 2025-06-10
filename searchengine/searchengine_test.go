package searchengine

import (
	"context"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"testing"

	"github.com/blevesearch/bleve/v2"
)

func init() {
	// 加载配置文件
	config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	_ = storage.InitStorage(context.Background())
}

func TestIndex(t *testing.T) {
	// 初始化搜索引擎组件
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	stats := Index.Stats()
	json, err := stats.MarshalJSON()
	t.Logf("索引统计信息: %v", string(json))
}

// TestIndexContent 测试索引内容的详细信息
func TestIndexContent(t *testing.T) {
	// 初始化搜索引擎组件
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// 获取索引的文档数量
	docCount, err := Index.DocCount()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("索引中的文档数量: %d", docCount)

	// 创建一个简单的匹配所有文档的查询
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 100             // 获取最多100个文档
	searchRequest.Fields = []string{"*"} // 获取所有字段

	// 执行搜索
	searchResult, err := Index.Search(searchRequest)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("搜索结果总数: %d", searchResult.Total)
	t.Logf("返回的文档数: %d", len(searchResult.Hits))

	// 显示每个文档的详细信息
	for i, hit := range searchResult.Hits {
		t.Logf("文档 %d:", i+1)
		t.Logf("  ID: %s", hit.ID)
		t.Logf("  Score: %f", hit.Score)
		t.Logf("  Fields: %v", hit.Fields)

		// 如果有高亮信息，也显示出来
		if len(hit.Fragments) > 0 {
			t.Logf("  Fragments: %v", hit.Fragments)
		}
		t.Logf("  ---")
	}
}

// TestSearchFunctionality 测试搜索功能
func TestSearchFunctionality(t *testing.T) {
	// 初始化搜索引擎组件
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// 测试不同类型的搜索查询
	searchTerms := []string{
		"博客", // 中文搜索
		"Go", // 英文搜索
		"程序", // 中文程序相关
		"title",
	}

	for _, term := range searchTerms {
		t.Logf("\n=== 搜索关键词: %s ===", term)

		// 创建查询
		query := bleve.NewMatchQuery(term)
		searchRequest := bleve.NewSearchRequest(query)
		searchRequest.Size = 10
		searchRequest.Fields = []string{"Title", "Content"}
		searchRequest.Highlight = bleve.NewHighlight()

		// 执行搜索
		searchResult, err := Index.Search(searchRequest)
		if err != nil {
			t.Errorf("搜索失败: %v", err)
			continue
		}

		t.Logf("找到 %d 个匹配结果", searchResult.Total)

		// 显示搜索结果
		for i, hit := range searchResult.Hits {
			t.Logf("结果 %d:", i+1)
			t.Logf("  文档ID: %s", hit.ID)
			t.Logf("  匹配分数: %f", hit.Score)

			if title, ok := hit.Fields["Title"]; ok {
				t.Logf("  标题: %s", title)
			}

			// 显示高亮片段
			if len(hit.Fragments) > 0 {
				t.Logf("  匹配片段:")
				for field, fragments := range hit.Fragments {
					t.Logf("    %s: %v", field, fragments)
				}
			}
			t.Logf("  ---")
		}
	}
}
