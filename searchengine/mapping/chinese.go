package mapping

import (
	"sparrow_blog_server/searchengine/tokenizer"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/token/length"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/registry"
)

// CreateChineseMapping 创建针对中文的索引映射
// 该函数定制了一个适合处理中文文本的搜索索引映射
// 返回值是定制好的索引映射和一个错误对象，如果创建过程中出现任何问题，就会返回相应的错误
func CreateChineseMapping() (mapping.IndexMapping, error) {
	// 1. 注册中文分词器
	// 这里注册了一个名为"chinese"的分词器，用于后续的文本分析
	if err := registry.RegisterTokenizer("chinese", func(config map[string]interface{}, cache *registry.Cache) (analysis.Tokenizer, error) {
		return tokenizer.NewChineseTokenizer(), nil
	}); err != nil {
		return nil, err
	}

	// 2. 创建索引映射
	// 这是 bleve 库用于定义索引结构的映射，我们将在其基础上添加自定义的分析器和字段映射
	indexMapping := bleve.NewIndexMapping()

	// 3. 注册自定义长度过滤器
	// "min_length"过滤器用于移除长度不符合要求的词汇，这里设定词汇长度必须在2到20个字符之间
	err := indexMapping.AddCustomTokenFilter("min_length", map[string]any{
		"type": length.Name,
		"min":  2.0,
		"max":  20.0,
	})
	if err != nil {
		return nil, err
	}

	// 4. 创建自定义分析器
	// 这里定义了一个名为"chinese_analyzer"的分析器，它使用了之前注册的中文分词器
	// 并结合了一些预处理步骤，如转换为小写和一个自定义的最小长度过滤器
	err = indexMapping.AddCustomAnalyzer("chinese_analyzer", map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": "chinese",
		"token_filters": []string{
			"to_lower",   // 小写转换
			"min_length", // 最小长度过滤
		},
	})
	if err != nil {
		return nil, err
	}

	// 5. 创建文档映射
	// 文档映射定义了索引中文档的结构，这里我们关注的是如何处理文档中的字段
	articleMapping := bleve.NewDocumentMapping()

	// 6. 配置字段使用自定义分析器
	// 对"Title"和"Content"字段应用之前定义的中文分析器，以优化中文文本的搜索
	titleField := bleve.NewTextFieldMapping()
	titleField.Analyzer = "chinese_analyzer"

	contentField := bleve.NewTextFieldMapping()
	contentField.Analyzer = "chinese_analyzer"

	// 7. 将字段映射添加到文档
	// 这一步将之前定义的字段映射到文档映射中，以便在索引时应用这些配置
	articleMapping.AddFieldMappingsAt("Title", titleField)
	articleMapping.AddFieldMappingsAt("Content", contentField)

	// 8. 将文档映射添加到索引映射
	// 最后，将文档映射添加到索引映射中，这样就完成了整个中文索引映射的创建
	indexMapping.AddDocumentMapping("Article", articleMapping)

	// 返回定制好的索引映射和nil错误，表示执行成功
	return indexMapping, nil
}
