package mapping

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/token/length"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/mapping"
)

// CreateChineseMapping 创建针对中文的索引映射
// 使用 bleve 内置的 unicode 分词器，支持中文、英文等多种语言
// 无需 CGO 依赖，更加稳定和轻量
func CreateChineseMapping() (mapping.IndexMapping, error) {
	// 1. 创建索引映射
	indexMapping := bleve.NewIndexMapping()

	// 2. 注册自定义长度过滤器
	// "min_length"过滤器用于移除长度不符合要求的词汇，这里设定词汇长度必须在1到20个字符之间
	err := indexMapping.AddCustomTokenFilter("min_length", map[string]any{
		"type": length.Name,
		"min":  1.0, // 中文单字也很有意义，所以最小长度设为1
		"max":  20.0,
	})
	if err != nil {
		return nil, err
	}

	// 3. 创建自定义分析器
	// 使用 bleve 内置的 unicode 分词器，它能很好地处理中文、英文等多种语言
	// unicode 分词器基于 Unicode 文本分割标准，在词边界处分割文本
	err = indexMapping.AddCustomAnalyzer("unicode_analyzer", map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": unicode.Name, // 使用内置的 unicode 分词器
		"token_filters": []string{
			"to_lower",   // 小写转换
			"min_length", // 最小长度过滤
		},
	})
	if err != nil {
		return nil, err
	}

	// 4. 创建默认文档映射
	defaultMapping := bleve.NewDocumentMapping()

	// 5. 配置字段使用 unicode 分析器
	// 对"Title"和"Content"字段应用 unicode 分析器，优化多语言文本的搜索
	titleField := bleve.NewTextFieldMapping()
	titleField.Analyzer = "unicode_analyzer"
	titleField.Store = true // 设置为存储，以便在搜索结果中返回字段内容
	titleField.Index = true // 确保字段被索引

	contentField := bleve.NewTextFieldMapping()
	contentField.Analyzer = "unicode_analyzer"
	contentField.Store = true // 设置为存储，以便在搜索结果中返回字段内容
	contentField.Index = true // 确保字段被索引

	// ID字段配置（用于精确匹配，不需要分析器）
	idField := bleve.NewTextFieldMapping()
	idField.Store = true
	idField.Index = true
	idField.Analyzer = "keyword" // 使用keyword分析器，不分词

	// ImgId字段配置（用于精确匹配，不需要分析器）
	imgIdField := bleve.NewTextFieldMapping()
	imgIdField.Store = true
	imgIdField.Index = true
	imgIdField.Analyzer = "keyword" // 使用keyword分析器，不分词

	// 6. 将字段映射添加到默认文档映射
	defaultMapping.AddFieldMappingsAt("ID", idField)
	defaultMapping.AddFieldMappingsAt("ImgId", imgIdField)
	defaultMapping.AddFieldMappingsAt("Title", titleField)
	defaultMapping.AddFieldMappingsAt("Content", contentField)

	// 启用动态映射，允许未明确定义的字段也被索引
	defaultMapping.Dynamic = true

	// 7. 设置默认文档映射
	// 使用默认映射而不是命名类型映射，这样所有文档都会使用这个映射
	indexMapping.DefaultMapping = defaultMapping

	// 设置默认类型（可选）
	indexMapping.TypeField = "_type"
	indexMapping.DefaultType = "_default"

	// 返回定制好的索引映射和nil错误，表示执行成功
	return indexMapping, nil
}
