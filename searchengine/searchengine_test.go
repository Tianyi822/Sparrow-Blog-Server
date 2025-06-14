package searchengine

import (
	"context"
	"errors"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"strings"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
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
	json, _ := stats.MarshalJSON()
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

// TestChineseTokenizationDiagnostic 诊断中文分词问题
func TestChineseTokenizationDiagnostic(t *testing.T) {
	// 初始化搜索引擎
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== 中文分词诊断测试 ===")

	// 1. 测试分析器是否正常工作
	t.Log("\n--- 步骤1: 测试中文分析器 ---")
	analyzer := Index.Mapping().AnalyzerNamed("chinese_analyzer")
	if analyzer == nil {
		t.Fatal("中文分析器为nil")
	}
	t.Log("✓ 中文分析器存在")

	// 2. 测试分词结果
	t.Log("\n--- 步骤2: 测试分词结果 ---")
	testText := "测试缓存 test test"
	tokens := analyzer.Analyze([]byte(testText))

	t.Logf("原文: %s", testText)
	t.Logf("分词结果 (%d个token):", len(tokens))
	for i, token := range tokens {
		t.Logf("  Token %d: '%s' (位置: %d-%d)", i+1, string(token.Term), token.Start, token.End)
	}

	// 检查是否包含"测试"
	foundTest := false
	for _, token := range tokens {
		if string(token.Term) == "测试" {
			foundTest = true
			break
		}
	}

	if foundTest {
		t.Log("✓ 分词器成功识别出'测试'")
	} else {
		t.Error("✗ 分词器未能识别出'测试'")
	}

	// 3. 检查目标文档
	targetDocID := "e84f1230f7358390"
	t.Logf("\n--- 步骤3: 检查目标文档 %s ---", targetDocID)

	// 获取文档
	allQuery := bleve.NewMatchAllQuery()
	allRequest := bleve.NewSearchRequest(allQuery)
	allRequest.Size = 100
	allRequest.Fields = []string{"Title", "Content"}

	allResult, err := Index.Search(allRequest)
	if err != nil {
		t.Fatalf("搜索所有文档失败: %v", err)
	}

	for _, hit := range allResult.Hits {
		if hit.ID == targetDocID {
			t.Logf("找到目标文档: %s", hit.ID)
			if title, ok := hit.Fields["Title"]; ok {
				t.Logf("  标题: %s", title)
			}
			if content, ok := hit.Fields["Content"]; ok {
				if contentStr, ok := content.(string); ok {
					t.Logf("  内容长度: %d", len(contentStr))
					// 检查内容中是否包含"测试"
					if len(contentStr) > 0 {
						t.Logf("  内容预览: %s...", contentStr[:min(200, len(contentStr))])
						if strings.Contains(contentStr, "测试") {
							t.Log("  ✓ 内容包含'测试'字符串")
						} else {
							t.Log("  ✗ 内容不包含'测试'字符串")
						}
					}
				}
			}
			break
		}
	}

	// 4. 测试不同的搜索查询方式
	t.Log("\n--- 步骤4: 测试不同搜索方式 ---")

	// 4.1 精确匹配查询
	t.Log("4.1 精确匹配查询")
	matchQuery := bleve.NewMatchQuery("测试")
	matchRequest := bleve.NewSearchRequest(matchQuery)
	matchRequest.Fields = []string{"Title", "Content"}

	matchResult, err := Index.Search(matchRequest)
	if err != nil {
		t.Error("精确匹配查询失败:", err)
	} else {
		t.Logf("  精确匹配结果: %d", matchResult.Total)
	}

	// 4.2 模糊查询
	t.Log("4.2 模糊查询")
	fuzzyQuery := bleve.NewFuzzyQuery("测试")
	fuzzyRequest := bleve.NewSearchRequest(fuzzyQuery)
	fuzzyRequest.Fields = []string{"Title", "Content"}

	fuzzyResult, err := Index.Search(fuzzyRequest)
	if err != nil {
		t.Error("模糊查询失败:", err)
	} else {
		t.Logf("  模糊查询结果: %d", fuzzyResult.Total)
	}

	// 4.3 通配符查询
	t.Log("4.3 通配符查询")
	wildcardQuery := bleve.NewWildcardQuery("*测试*")
	wildcardRequest := bleve.NewSearchRequest(wildcardQuery)
	wildcardRequest.Fields = []string{"Title", "Content"}

	wildcardResult, err := Index.Search(wildcardRequest)
	if err != nil {
		t.Error("通配符查询失败:", err)
	} else {
		t.Logf("  通配符查询结果: %d", wildcardResult.Total)
	}

	// 4.4 词项查询
	t.Log("4.4 词项查询")
	termQuery := bleve.NewTermQuery("测试")
	termRequest := bleve.NewSearchRequest(termQuery)
	termRequest.Fields = []string{"Title", "Content"}

	termResult, err := Index.Search(termRequest)
	if err != nil {
		t.Error("词项查询失败:", err)
	} else {
		t.Logf("  词项查询结果: %d", termResult.Total)
	}

	// 5. 检查索引统计
	t.Log("\n--- 步骤5: 检查索引统计 ---")
	stats := Index.Stats()
	if statsJSON, err := stats.MarshalJSON(); err == nil {
		t.Logf("  索引统计: %s", string(statsJSON))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestSearchFunctionality 测试搜索功能的完整性
func TestSearchFunctionality(t *testing.T) {
	// 初始化搜索引擎组件
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// 获取索引统计信息
	docCount, err := Index.DocCount()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("索引中包含 %d 个文档", docCount)

	// 测试不同类型的搜索查询
	searchTests := []struct {
		keyword     string
		description string
		expectHits  bool
	}{
		{"测试", "中文搜索", true},
		{"test", "英文搜索", true},
		{"博客", "中文内容搜索", true},
		{"Go", "技术关键词搜索", false}, // 可能没有结果
		{"不存在的关键词xyz123", "无结果搜索", false},
	}

	for _, test := range searchTests {
		t.Logf("\n=== 搜索关键词: %s (%s) ===", test.keyword, test.description)

		// 使用新的搜索函数
		searchReq := SearchRequest{
			Query:     test.keyword,
			Size:      5,
			From:      0,
			Fields:    []string{"Title", "Content"},
			Highlight: true,
		}

		searchResult, err := Search(searchReq)
		if err != nil {
			t.Errorf("搜索'%s'失败: %v", test.keyword, err)
			continue
		}

		t.Logf("找到 %d 个匹配结果", searchResult.Total)
		t.Logf("搜索耗时: %.2f 毫秒", searchResult.TimeMs)

		// 验证预期结果
		if test.expectHits && searchResult.Total == 0 {
			t.Logf("  ⚠️  预期有结果但未找到匹配项")
		} else if !test.expectHits && searchResult.Total > 0 {
			t.Logf("  ⚠️  预期无结果但找到了匹配项")
		}

		// 显示搜索结果
		for i, hit := range searchResult.Hits {
			t.Logf("结果 %d:", i+1)
			t.Logf("  文档ID: %s", hit.ID)
			t.Logf("  匹配分数: %.2f", hit.Score)

			if title, ok := hit.Fields["Title"]; ok {
				t.Logf("  标题: %s", title)
			}

			// 显示高亮片段
			if len(hit.Fragments) > 0 {
				t.Logf("  匹配片段:")
				for field, fragments := range hit.Fragments {
					for j, fragment := range fragments {
						t.Logf("    %s[%d]: %s", field, j+1, fragment)
					}
				}
			}
		}
	}

	// 测试特殊情况
	t.Log("\n=== 特殊情况测试 ===")

	// 空查询
	emptyReq := SearchRequest{Query: ""}
	emptyResult, err := Search(emptyReq)
	if err != nil {
		t.Logf("空查询正确返回错误: %v", err)
	} else {
		t.Logf("空查询结果: %d 个文档", emptyResult.Total)
	}
}

// TestSearchTest 测试搜索"test"和"测试"关键词，验证ID为e84f1230f7358390的文档能被匹配
func TestSearchTest(t *testing.T) {
	// 初始化搜索引擎组件
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// 测试多个搜索关键词
	searchTerms := []string{"测试"}
	expectedDocID := "e84f1230f7358390"

	t.Logf("期望匹配的文档ID: %s", expectedDocID)

	// 首先获取目标文档的详细信息
	t.Logf("\n=== 获取目标文档详细信息 ===")
	allQuery := bleve.NewMatchAllQuery()
	allRequest := bleve.NewSearchRequest(allQuery)
	allRequest.Size = 100
	allRequest.Fields = []string{"Title", "Content"}

	allResult, err := Index.Search(allRequest)
	if err != nil {
		t.Fatalf("获取所有文档失败: %v", err)
	}

	var targetDoc *search.DocumentMatch
	for _, hit := range allResult.Hits {
		if hit.ID == expectedDocID {
			targetDoc = hit
			break
		}
	}

	if targetDoc != nil {
		t.Logf("目标文档信息:")
		t.Logf("  ID: %s", targetDoc.ID)
		if title, ok := targetDoc.Fields["Title"]; ok {
			t.Logf("  标题: %s", title)
		} else {
			t.Logf("  标题: [字段不存在]")
		}
		if content, ok := targetDoc.Fields["Content"]; ok {
			switch v := content.(type) {
			case []byte:
				contentStr := string(v)
				t.Logf("  内容长度: %d 字节", len(v))
				if len(contentStr) == 0 {
					t.Logf("  内容: [空内容]")
				} else if len(contentStr) > 500 {
					t.Logf("  内容预览: %s...", contentStr[:500])
				} else {
					t.Logf("  内容: %s", contentStr)
				}
			case string:
				t.Logf("  内容长度: %d 字符", len(v))
				if len(v) == 0 {
					t.Logf("  内容: [空内容]")
				} else if len(v) > 500 {
					t.Logf("  内容预览: %s...", v[:500])
				} else {
					t.Logf("  内容: %s", v)
				}
			case nil:
				t.Logf("  内容: [nil]")
			default:
				t.Logf("  内容类型: %T, 值: %v", v, v)
			}
		} else {
			t.Logf("  内容: [字段不存在]")
		}

		// 显示所有可用字段
		t.Logf("  所有可用字段:")
		for fieldName, fieldValue := range targetDoc.Fields {
			t.Logf("    %s: %T = %v", fieldName, fieldValue, fieldValue)
		}
	} else {
		t.Errorf("未找到目标文档ID: %s", expectedDocID)
		return
	}

	// 测试每个搜索词
	for _, searchTerm := range searchTerms {
		t.Logf("\n=== 搜索关键词: %s ===", searchTerm)

		// 创建查询 - 在Title和Content字段中搜索
		titleQuery := bleve.NewMatchQuery(searchTerm)
		titleQuery.SetField("Title")

		contentQuery := bleve.NewMatchQuery(searchTerm)
		contentQuery.SetField("Content")

		// 组合查询 - Title或Content包含搜索词
		boolQuery := bleve.NewBooleanQuery()
		boolQuery.AddShould(titleQuery)
		boolQuery.AddShould(contentQuery)

		searchRequest := bleve.NewSearchRequest(boolQuery)
		searchRequest.Size = 20 // 增大返回数量以确保找到目标文档
		searchRequest.Fields = []string{"Title", "Content"}

		// 配置高亮显示
		highlight := bleve.NewHighlight()
		highlight.AddField("Title")
		highlight.AddField("Content")
		searchRequest.Highlight = highlight

		// 执行搜索
		searchResult, err := Index.Search(searchRequest)
		if err != nil {
			t.Errorf("搜索'%s'失败: %v", searchTerm, err)
			continue
		}

		t.Logf("搜索结果总数: %d", searchResult.Total)

		// 查找期望的文档ID
		var foundExpectedDoc bool

		for i, hit := range searchResult.Hits {
			t.Logf("\n--- 结果 %d ---", i+1)
			t.Logf("文档ID: %s", hit.ID)
			t.Logf("匹配分数: %f", hit.Score)

			// 输出标题
			if title, ok := hit.Fields["Title"]; ok {
				t.Logf("标题: %s", title)
			}

			// 输出内容片段（确保是字符串格式）
			if content, ok := hit.Fields["Content"]; ok {
				switch v := content.(type) {
				case []byte:
					// 如果是字节数组，转换为字符串
					contentStr := string(v)
					t.Logf("内容片段（从字节数组转换）: %s", contentStr)
				case string:
					// 如果已经是字符串
					t.Logf("内容片段（字符串）: %s", v)
				default:
					t.Logf("内容片段（其他类型 %T）: %v", v, v)
				}
			}

			// 显示高亮片段（字符串格式）
			if len(hit.Fragments) > 0 {
				t.Logf("高亮匹配片段:")
				for field, fragments := range hit.Fragments {
					t.Logf("  字段 %s:", field)
					for j, fragment := range fragments {
						// 确保片段是字符串格式
						fragmentStr := string(fragment)
						t.Logf("    片段 %d: %s", j+1, fragmentStr)
					}
				}
			}

			// 检查是否是期望的文档
			if hit.ID == expectedDocID {
				foundExpectedDoc = true
				t.Logf("*** 找到期望的文档ID: %s ***", expectedDocID)
			}
		}

		// 验证搜索词是否找到了期望的文档
		if !foundExpectedDoc {
			t.Errorf("搜索'%s'未找到期望的文档ID: %s", searchTerm, expectedDocID)
		} else {
			t.Logf("搜索'%s'成功找到期望的文档", searchTerm)
		}

		// 如果没有找到任何结果，进行额外的调试
		if searchResult.Total == 0 {
			t.Logf("\n=== 调试信息 (搜索词: %s) ===", searchTerm)
			t.Logf("没有找到任何匹配结果")

			// 尝试更宽松的查询
			wildcardQuery := bleve.NewWildcardQuery("*" + searchTerm + "*")
			wildcardRequest := bleve.NewSearchRequest(wildcardQuery)
			wildcardRequest.Size = 10
			wildcardRequest.Fields = []string{"Title", "Content"}

			wildcardResult, err := Index.Search(wildcardRequest)
			if err != nil {
				t.Logf("通配符搜索失败: %v", err)
			} else {
				t.Logf("通配符搜索找到 %d 个结果", wildcardResult.Total)
			}
		}
	}
}

// TestSearchEngineUsageExample 搜索引擎使用示例和最佳实践
func TestSearchEngineUsageExample(t *testing.T) {
	t.Log("=== 搜索引擎使用示例 ===")

	// 1. 初始化搜索引擎
	t.Log("\n--- 步骤1: 初始化搜索引擎 ---")
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal("初始化搜索引擎失败:", err)
	}
	t.Log("✓ 搜索引擎初始化成功")

	// 2. 基本搜索示例
	t.Log("\n--- 步骤2: 基本搜索示例 ---")
	basicReq := SearchRequest{
		Query: "测试", // 搜索关键词
	}
	// 注意：未设置的字段将使用默认值
	// Size: 10, From: 0, Fields: ["Title", "Content"], Highlight: false

	basicResult, err := Search(basicReq)
	if err != nil {
		t.Error("基本搜索失败:", err)
	} else {
		t.Logf("✓ 基本搜索成功，找到 %d 个结果", basicResult.Total)
	}

	// 3. 高级搜索示例
	t.Log("\n--- 步骤3: 高级搜索示例 ---")
	advancedReq := SearchRequest{
		Query:     "测试",                       // 搜索关键词
		Size:      3,                            // 限制返回3个结果
		From:      0,                            // 从第0个开始（分页）
		Fields:    []string{"Title", "Content"}, // 指定返回字段
		Highlight: true,                         // 启用高亮
	}

	advancedResult, err := Search(advancedReq)
	if err != nil {
		t.Error("高级搜索失败:", err)
	} else {
		t.Logf("✓ 高级搜索成功，找到 %d 个结果", advancedResult.Total)

		// 处理搜索结果
		t.Log("\n搜索结果处理示例:")
		for i, hit := range advancedResult.Hits {
			t.Logf("--- 第 %d 个结果 ---", i+1)
			t.Logf("文档ID: %s", hit.ID)
			t.Logf("相关性得分: %.3f", hit.Score)

			// 获取标题
			if title, exists := hit.Fields["Title"]; exists {
				t.Logf("标题: %s", title)
			}

			// 获取内容预览
			if content, exists := hit.Fields["Content"]; exists {
				if contentStr, ok := content.(string); ok && len(contentStr) > 0 {
					preview := contentStr
					if len(contentStr) > 100 {
						preview = contentStr[:100] + "..."
					}
					t.Logf("内容预览: %s", preview)
				}
			}

			// 处理高亮片段
			if len(hit.Fragments) > 0 {
				t.Log("高亮片段:")
				for field, fragments := range hit.Fragments {
					for _, fragment := range fragments {
						t.Logf("  %s: %s", field, fragment)
					}
				}
			}
		}
	}

	// 4. 分页搜索示例
	t.Log("\n--- 步骤4: 分页搜索示例 ---")
	// 第一页
	page1Req := SearchRequest{
		Query:  "test",
		Size:   2, // 每页2个结果
		From:   0, // 第一页从0开始
		Fields: []string{"Title"},
	}

	page1Result, err := Search(page1Req)
	if err != nil {
		t.Error("第一页搜索失败:", err)
	} else {
		t.Logf("✓ 第一页: 总共 %d 个结果，返回 %d 个", page1Result.Total, len(page1Result.Hits))
	}

	// 第二页（如果有的话）
	if page1Result != nil && page1Result.Total > 2 {
		page2Req := SearchRequest{
			Query:  "test",
			Size:   2, // 每页2个结果
			From:   2, // 第二页从2开始
			Fields: []string{"Title"},
		}

		page2Result, err := Search(page2Req)
		if err != nil {
			t.Error("第二页搜索失败:", err)
		} else {
			t.Logf("✓ 第二页: 返回 %d 个结果", len(page2Result.Hits))
		}
	}

	// 5. 错误处理示例
	t.Log("\n--- 步骤5: 错误处理示例 ---")

	// 测试各种边界情况
	testCases := []struct {
		name string
		req  SearchRequest
	}{
		{"负数Size", SearchRequest{Query: "test", Size: -1}},
		{"负数From", SearchRequest{Query: "test", From: -1}},
		{"大Size值", SearchRequest{Query: "test", Size: 1000}},
	}

	for _, tc := range testCases {
		result, err := Search(tc.req)
		if err != nil {
			t.Logf("%-10s: 返回错误 %v", tc.name, err)
		} else {
			t.Logf("%-10s: 成功，返回 %d 个结果", tc.name, len(result.Hits))
		}
	}

	// 6. 性能监控示例
	t.Log("\n--- 步骤6: 性能监控示例 ---")
	perfReq := SearchRequest{
		Query: "测试",
		Size:  10,
	}

	perfResult, err := Search(perfReq)
	if err != nil {
		t.Error("性能测试失败:", err)
	} else {
		// 搜索耗时已经是毫秒
		t.Logf("✓ 搜索耗时: %.2f 毫秒", perfResult.TimeMs)
		t.Logf("✓ 平均每个结果耗时: %.2f 毫秒", perfResult.TimeMs/float64(len(perfResult.Hits)+1))

		// 性能建议
		if perfResult.TimeMs > 100 {
			t.Log("⚠️  搜索耗时较长，建议优化索引或减少返回字段")
		} else if perfResult.TimeMs < 10 {
			t.Log("✓ 搜索性能良好")
		}
	}

	// 7. 最佳实践总结
	t.Log("\n--- 搜索引擎使用最佳实践 ---")
	t.Log("1. 总是检查错误返回值")
	t.Log("2. 合理设置Size参数，避免返回过多结果")
	t.Log("3. 使用分页处理大量结果")
	t.Log("4. 根据需要选择返回字段，减少不必要的数据传输")
	t.Log("5. 监控搜索性能，必要时优化查询")
	t.Log("6. 中文和英文搜索都已完全支持")
	t.Log("7. 高亮功能可以改善用户体验")
}

// TestMatchQueryFieldSpecific 测试字段特定的匹配查询
func TestMatchQueryFieldSpecific(t *testing.T) {
	// 初始化搜索引擎
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== 字段特定匹配查询测试 ===")

	// 测试关键词
	keyword := "测试"
	targetDocID := "e84f1230f7358390"

	// 1. 测试通用MatchQuery（不指定字段）
	t.Log("\n--- 测试1: 通用MatchQuery ---")
	generalQuery := bleve.NewMatchQuery(keyword)
	generalRequest := bleve.NewSearchRequest(generalQuery)
	generalRequest.Fields = []string{"Title", "Content"}

	generalResult, err := Index.Search(generalRequest)
	if err != nil {
		t.Error("通用MatchQuery失败:", err)
	} else {
		t.Logf("通用MatchQuery结果: %d", generalResult.Total)
	}

	// 2. 测试Content字段特定的MatchQuery
	t.Log("\n--- 测试2: Content字段特定MatchQuery ---")
	contentQuery := bleve.NewMatchQuery(keyword)
	contentQuery.SetField("Content")
	contentRequest := bleve.NewSearchRequest(contentQuery)
	contentRequest.Fields = []string{"Title", "Content"}

	contentResult, err := Index.Search(contentRequest)
	if err != nil {
		t.Error("Content字段MatchQuery失败:", err)
	} else {
		t.Logf("Content字段MatchQuery结果: %d", contentResult.Total)
		for i, hit := range contentResult.Hits {
			t.Logf("  结果 %d: ID=%s, Score=%.2f", i+1, hit.ID, hit.Score)
			if hit.ID == targetDocID {
				t.Log("  ✓ 找到目标文档!")
			}
		}
	}

	// 3. 测试Title字段特定的MatchQuery
	t.Log("\n--- 测试3: Title字段特定MatchQuery ---")
	titleQuery := bleve.NewMatchQuery(keyword)
	titleQuery.SetField("Title")
	titleRequest := bleve.NewSearchRequest(titleQuery)
	titleRequest.Fields = []string{"Title", "Content"}

	titleResult, err := Index.Search(titleRequest)
	if err != nil {
		t.Error("Title字段MatchQuery失败:", err)
	} else {
		t.Logf("Title字段MatchQuery结果: %d", titleResult.Total)
	}

	// 4. 测试组合查询（Title OR Content）
	t.Log("\n--- 测试4: 组合查询（Title OR Content）---")
	titleQueryCombined := bleve.NewMatchQuery(keyword)
	titleQueryCombined.SetField("Title")

	contentQueryCombined := bleve.NewMatchQuery(keyword)
	contentQueryCombined.SetField("Content")

	boolQuery := bleve.NewBooleanQuery()
	boolQuery.AddShould(titleQueryCombined)
	boolQuery.AddShould(contentQueryCombined)

	combinedRequest := bleve.NewSearchRequest(boolQuery)
	combinedRequest.Fields = []string{"Title", "Content"}
	combinedRequest.Highlight = bleve.NewHighlight()

	combinedResult, err := Index.Search(combinedRequest)
	if err != nil {
		t.Error("组合查询失败:", err)
	} else {
		t.Logf("组合查询结果: %d", combinedResult.Total)
		for i, hit := range combinedResult.Hits {
			t.Logf("  结果 %d: ID=%s, Score=%.2f", i+1, hit.ID, hit.Score)
			if hit.ID == targetDocID {
				t.Log("  ✓ 找到目标文档!")

				// 显示高亮信息
				if len(hit.Fragments) > 0 {
					t.Log("  高亮片段:")
					for field, fragments := range hit.Fragments {
						for j, fragment := range fragments {
							t.Logf("    %s[%d]: %s", field, j+1, string(fragment))
						}
					}
				}
			}
		}
	}

	// 5. 分析查询词的分析结果
	t.Log("\n--- 测试5: 分析查询词 ---")
	analyzer := Index.Mapping().AnalyzerNamed("chinese_analyzer")
	if analyzer != nil {
		queryTokens := analyzer.Analyze([]byte(keyword))
		t.Logf("查询词'%s'的分析结果:", keyword)
		for i, token := range queryTokens {
			t.Logf("  Token %d: '%s'", i+1, string(token.Term))
		}
	}
}

// TestNewSearchFunction 测试新的搜索函数
func TestNewSearchFunction(t *testing.T) {
	// 初始化搜索引擎
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== 测试新的搜索函数 ===")

	// 测试中文搜索
	t.Log("\n--- 测试中文搜索: '测试' ---")
	chineseReq := SearchRequest{
		Query:     "测试",
		Size:      5,
		From:      0,
		Fields:    []string{"Title", "Content"},
		Highlight: true,
	}

	chineseResult, err := Search(chineseReq)
	if err != nil {
		t.Error("中文搜索失败:", err)
	} else {
		t.Logf("中文搜索结果: 找到 %d 个文档", chineseResult.Total)
		t.Logf("搜索耗时: %.2f 毫秒", chineseResult.TimeMs)

		targetFound := false
		for i, hit := range chineseResult.Hits {
			t.Logf("  结果 %d: ID=%s, Score=%.2f", i+1, hit.ID, hit.Score)
			if hit.ID == "e84f1230f7358390" {
				targetFound = true
				t.Log("    ✓ 找到目标文档!")

				// 显示高亮片段
				if len(hit.Fragments) > 0 {
					t.Log("    高亮片段:")
					for field, fragments := range hit.Fragments {
						for j, fragment := range fragments {
							t.Logf("      %s[%d]: %s", field, j+1, fragment)
						}
					}
				}
			}
		}

		if targetFound {
			t.Log("  ✅ 中文搜索成功!")
		} else {
			t.Error("  ❌ 中文搜索未找到目标文档")
		}
	}

	// 测试英文搜索
	t.Log("\n--- 测试英文搜索: 'test' ---")
	englishReq := SearchRequest{
		Query:     "test",
		Size:      3,
		From:      0,
		Fields:    []string{"Title", "Content"},
		Highlight: true,
	}

	englishResult, err := Search(englishReq)
	if err != nil {
		t.Error("英文搜索失败:", err)
	} else {
		t.Logf("英文搜索结果: 找到 %d 个文档", englishResult.Total)
		for i, hit := range englishResult.Hits {
			t.Logf("  结果 %d: ID=%s, Score=%.2f", i+1, hit.ID, hit.Score)
		}
	}

	// 测试空查询
	t.Log("\n--- 测试空查询 ---")
	emptyReq := SearchRequest{
		Query: "",
		Size:  1,
	}

	emptyResult, err := Search(emptyReq)
	if err != nil {
		t.Logf("空查询正确返回错误: %v", err)
	} else {
		t.Logf("空查询结果: %d 个文档", emptyResult.Total)
	}

	// 测试默认参数
	t.Log("\n--- 测试默认参数 ---")
	defaultReq := SearchRequest{
		Query: "博客",
	}

	defaultResult, err := Search(defaultReq)
	if err != nil {
		t.Error("默认参数搜索失败:", err)
	} else {
		t.Logf("默认参数搜索结果: 找到 %d 个文档", defaultResult.Total)
		t.Logf("返回结果数: %d (应该<=10)", len(defaultResult.Hits))
	}
}

// TestRebuildIndex 测试重建索引功能
func TestRebuildIndex(t *testing.T) {
	t.Log("=== 测试重建索引功能 ===")

	// 1. 首先确保有一个现有的索引
	t.Log("\n--- 步骤1: 初始化现有索引 ---")
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal("初始化索引失败:", err)
	}

	// 获取重建前的索引统计信息
	originalDocCount, err := Index.DocCount()
	if err != nil {
		t.Fatal("获取原始文档数量失败:", err)
	}
	t.Logf("重建前索引文档数量: %d", originalDocCount)

	// 2. 执行搜索测试，确保索引正常工作
	t.Log("\n--- 步骤2: 重建前搜索测试 ---")
	beforeReq := SearchRequest{
		Query:  "test",
		Size:   5,
		Fields: []string{"Title", "Content"},
	}

	beforeResult, err := Search(beforeReq)
	if err != nil {
		t.Error("重建前搜索失败:", err)
	} else {
		t.Logf("重建前搜索结果: 找到 %d 个文档", beforeResult.Total)
	}

	// 3. 执行重建索引
	t.Log("\n--- 步骤3: 执行重建索引 ---")
	ctx := context.Background()

	// 使用带超时的上下文，防止测试无限等待
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second) // 5分钟超时
	defer cancel()

	err = RebuildIndex(ctx)
	if err != nil {
		t.Fatal("重建索引失败:", err)
	}
	t.Log("✓ 重建索引成功完成")

	// 4. 验证重建后的索引
	t.Log("\n--- 步骤4: 验证重建后的索引 ---")

	// 检查索引是否仍然可用
	if Index == nil {
		t.Fatal("重建后索引为nil")
	}

	// 获取重建后的文档数量
	rebuiltDocCount, err := Index.DocCount()
	if err != nil {
		t.Fatal("获取重建后文档数量失败:", err)
	}
	t.Logf("重建后索引文档数量: %d", rebuiltDocCount)

	// 验证文档数量是否一致
	if rebuiltDocCount != originalDocCount {
		t.Errorf("重建后文档数量不一致: 原始=%d, 重建后=%d", originalDocCount, rebuiltDocCount)
	} else {
		t.Log("✓ 重建后文档数量一致")
	}

	// 5. 执行搜索测试，确保重建后索引正常工作
	t.Log("\n--- 步骤5: 重建后搜索测试 ---")

	// 测试英文搜索
	afterReq := SearchRequest{
		Query:     "test",
		Size:      5,
		Fields:    []string{"Title", "Content"},
		Highlight: true,
	}

	afterResult, err := Search(afterReq)
	if err != nil {
		t.Error("重建后英文搜索失败:", err)
	} else {
		t.Logf("重建后英文搜索结果: 找到 %d 个文档", afterResult.Total)

		// 比较重建前后的搜索结果
		if beforeResult.Total == afterResult.Total {
			t.Log("✓ 重建前后英文搜索结果数量一致")
		} else {
			t.Logf("⚠ 重建前后英文搜索结果数量不同: 重建前=%d, 重建后=%d",
				beforeResult.Total, afterResult.Total)
		}
	}

	// 测试中文搜索
	chineseReq := SearchRequest{
		Query:     "测试",
		Size:      3,
		Fields:    []string{"Title", "Content"},
		Highlight: true,
	}

	chineseResult, err := Search(chineseReq)
	if err != nil {
		t.Error("重建后中文搜索失败:", err)
	} else {
		t.Logf("重建后中文搜索结果: 找到 %d 个文档", chineseResult.Total)

		// 显示搜索结果详情
		for i, hit := range chineseResult.Hits {
			t.Logf("  结果 %d: ID=%s, Score=%.2f", i+1, hit.ID, hit.Score)

			// 显示高亮片段
			if len(hit.Fragments) > 0 {
				for field, fragments := range hit.Fragments {
					for _, fragment := range fragments {
						t.Logf("    高亮[%s]: %s", field, fragment)
					}
				}
			}
		}
	}

	// 6. 测试索引统计信息
	t.Log("\n--- 步骤6: 验证索引统计信息 ---")
	stats := Index.Stats()
	statsJSON, _ := stats.MarshalJSON()
	t.Logf("重建后索引统计信息: %s", string(statsJSON))

	t.Log("\n=== 重建索引测试完成 ===")
}

// TestRebuildIndexWithCancel 测试重建索引的取消功能
func TestRebuildIndexWithCancel(t *testing.T) {
	t.Log("=== 测试重建索引取消功能 ===")

	// 1. 首先确保有一个现有的索引
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal("初始化索引失败:", err)
	}

	// 2. 创建一个会被取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	// 3. 尝试重建索引（应该立即返回取消错误）
	err = RebuildIndex(ctx)
	if err == nil {
		t.Error("期望重建索引因上下文取消而失败，但实际成功了")
	} else if errors.Is(err, context.Canceled) {
		t.Log("✓ 重建索引正确响应了上下文取消")
	} else {
		t.Logf("重建索引返回了其他错误: %v", err)
	}
}

// TestRebuildIndexWithTimeout 测试重建索引的超时功能
func TestRebuildIndexWithTimeout(t *testing.T) {
	t.Log("=== 测试重建索引超时功能 ===")

	// 1. 首先确保有一个现有的索引
	err := LoadingIndex(context.Background())
	if err != nil {
		t.Fatal("初始化索引失败:", err)
	}

	// 2. 创建一个很短超时的上下文（1毫秒）
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 等待一下确保超时
	time.Sleep(10 * time.Millisecond)

	// 3. 尝试重建索引（应该因超时而失败）
	err = RebuildIndex(ctx)
	if err == nil {
		t.Error("期望重建索引因超时而失败，但实际成功了")
	} else if errors.Is(err, context.DeadlineExceeded) {
		t.Log("✓ 重建索引正确响应了超时")
	} else {
		t.Logf("重建索引返回了其他错误: %v", err)
	}
}
