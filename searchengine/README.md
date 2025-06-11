# 搜索引擎使用指南

本搜索引擎基于 Bleve 构建，完全支持中英文搜索，具有高性能和易用性的特点。

## 功能特性

- ✅ **完整的中文支持** - 使用 jieba 分词器，支持中文搜索
- ✅ **英文搜索支持** - 原生支持英文关键词搜索  
- ✅ **高亮显示** - 自动高亮匹配的关键词
- ✅ **分页支持** - 支持分页查询大量结果
- ✅ **多字段搜索** - 同时在标题和内容中搜索
- ✅ **高性能** - 基于倒排索引，毫秒级响应

## 快速开始

### 1. 初始化搜索引擎

```go
import "sparrow_blog_server/searchengine"

// 初始化搜索引擎（通常在应用启动时调用一次）
err := searchengine.LoadingIndex(context.Background())
if err != nil {
    log.Fatal("搜索引擎初始化失败:", err)
}
```

### 2. 基本搜索

```go
// 创建搜索请求
req := searchengine.SearchRequest{
    Query: "测试",  // 搜索关键词（支持中英文）
}

// 执行搜索（使用默认参数）
result, err := searchengine.Search(req)
if err != nil {
    log.Printf("搜索失败: %v", err)
    return
}

// 处理搜索结果
fmt.Printf("找到 %d 个结果\n", result.Total)
for i, hit := range result.Hits {
    fmt.Printf("结果 %d: %s (分数: %.2f)\n", i+1, hit.ID, hit.Score)
}
```

### 3. 高级搜索

```go
// 高级搜索请求
req := searchengine.SearchRequest{
    Query:     "Go 语言",              // 搜索关键词
    Size:      20,                    // 返回结果数量
    From:      0,                     // 分页偏移量
    Fields:    []string{FieldTitle, FieldContent}, // 返回字段
    Highlight: true,                  // 启用高亮
}

result, err := searchengine.Search(req)
if err != nil {
    log.Printf("搜索失败: %v", err)
    return
}

// 处理搜索结果和高亮
for _, hit := range result.Hits {
    // 获取标题
    if title, exists := hit.Fields[FieldTitle]; exists {
        fmt.Printf("标题: %s\n", title)
    }
    
    // 获取内容预览
    if content, exists := hit.Fields[FieldContent]; exists {
        if contentStr, ok := content.(string); ok {
            preview := contentStr
            if len(contentStr) > 100 {
                preview = contentStr[:100] + "..."
            }
            fmt.Printf("内容: %s\n", preview)
        }
    }
    
    // 显示高亮片段
    if len(hit.Fragments) > 0 {
        fmt.Println("匹配片段:")
        for field, fragments := range hit.Fragments {
            for _, fragment := range fragments {
                fmt.Printf("  %s: %s\n", field, fragment)
            }
        }
    }
}
```

### 4. 分页搜索

```go
// 第一页
page1 := searchengine.SearchRequest{
    Query: "博客",
    Size:  10,     // 每页10个结果
    From:  0,      // 第一页从0开始
}

result1, err := searchengine.Search(page1)
if err != nil {
    log.Printf("搜索失败: %v", err)
    return
}

// 第二页
page2 := searchengine.SearchRequest{
    Query: "博客",
    Size:  10,     // 每页10个结果  
    From:  10,     // 第二页从10开始
}

result2, err := searchengine.Search(page2)
// 处理结果...
```

## API 参考

### 字段常量

为了避免硬编码，搜索引擎提供了以下字段常量：

```go
const (
    FieldID      = "ID"      // 文档ID字段
    FieldTitle   = "Title"   // 标题字段
    FieldContent = "Content" // 内容字段
)

// 默认搜索字段
var DefaultSearchFields = []string{FieldTitle, FieldContent}
```

### SearchRequest 结构

```go
type SearchRequest struct {
    Query      string   // 搜索关键词（必填）
    Size       int      // 返回结果数量，默认10
    From       int      // 分页偏移量，默认0
    Fields     []string // 返回字段，默认为DefaultSearchFields
    Highlight  bool     // 是否启用高亮，默认false
}
```

### SearchResponse 结构

```go
type SearchResponse struct {
    Total   uint64                     // 总结果数
    Hits    []*search.DocumentMatch    // 搜索结果
    TimeMs  float64                    // 搜索耗时（毫秒）
}
```

### DocumentMatch 结构（来自 Bleve）

每个搜索结果包含：
- `ID` - 文档唯一标识
- `Score` - 相关性分数（0-1之间，越高越相关）
- `Fields` - 返回的字段数据
- `Fragments` - 高亮片段（如果启用高亮）

## 性能优化建议

### 1. 合理设置参数

```go
// ✅ 推荐：限制返回数量
req := searchengine.SearchRequest{
    Query: "关键词",
    Size:  20,  // 不要设置过大的值
}

// ❌ 不推荐：返回过多结果
req := searchengine.SearchRequest{
    Query: "关键词", 
    Size:  1000,  // 可能影响性能
}
```

### 2. 选择必要的字段

```go
// ✅ 推荐：只返回需要的字段
req := searchengine.SearchRequest{
    Query:  "关键词",
    Fields: []string{FieldTitle},  // 只返回标题
}

// ❌ 不推荐：返回不必要的大字段
req := searchengine.SearchRequest{
    Query:  "关键词",
    Fields: []string{FieldTitle, FieldContent},  // Content字段可能很大
}
```

### 3. 监控性能

```go
result, err := searchengine.Search(req)
if err != nil {
    return err
}

// 监控搜索耗时（已经是毫秒）
if result.TimeMs > 100 {
    log.Printf("⚠️  搜索耗时较长: %.2f毫秒", result.TimeMs)
}
```

## 错误处理

```go
result, err := searchengine.Search(req)
if err != nil {
    // 处理搜索错误
    log.Printf("搜索失败: %v", err)
    
    // 根据错误类型进行不同处理
    switch {
    case strings.Contains(err.Error(), "empty query"):
        // 空查询错误
        return errors.New("搜索关键词不能为空")
    default:
        // 其他错误
        return errors.New("搜索服务暂时不可用")
    }
}

// 检查结果数量
if result.Total == 0 {
    log.Println("没有找到匹配的结果")
    return
}
```

## 最佳实践

1. **总是检查错误** - 搜索可能因为各种原因失败
2. **合理分页** - 使用 `Size` 和 `From` 参数实现分页
3. **控制返回字段** - 只请求需要的字段以提高性能
4. **启用高亮** - 使用高亮功能改善用户体验
5. **监控性能** - 记录搜索耗时，及时发现性能问题
6. **验证输入** - 在搜索前验证关键词的合法性

## 搜索示例

### 中文搜索
```go
// 中文关键词搜索
req := searchengine.SearchRequest{
    Query: "人工智能",
    Size:  10,
    Highlight: true,
}
```

### 英文搜索  
```go
// 英文关键词搜索
req := searchengine.SearchRequest{
    Query: "machine learning",
    Size:  10,
    Highlight: true,
}
```

### 混合搜索
```go
// 中英文混合搜索
req := searchengine.SearchRequest{
    Query: "Go 语言 tutorial",
    Size:  15,
    Highlight: true,
}
```

## 注意事项

- 搜索引擎在应用启动时自动建立索引
- 中文分词使用 jieba 算法，支持智能分词
- 搜索结果按相关性分数排序
- 高亮片段会自动截取匹配内容的上下文
- 空查询会返回0个结果（不会报错）

---

如需更多技术支持，请参考测试用例文件 `searchengine_test.go` 中的详细示例。 