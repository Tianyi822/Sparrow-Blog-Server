package webrouter

import (
	"net/url"

	"sparrow_blog_server/internal/services/adminservices"
	"sparrow_blog_server/internal/services/webservice"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/routers/resp"
	"sparrow_blog_server/searchengine"

	"github.com/gin-gonic/gin"
)

func getSysStatus(ctx *gin.Context) {
	if config.User.Username == "" {
		resp.Err(ctx, "服务状态异常，请检查配置文件", nil)
		return
	}

	resp.Ok(ctx, "获取成功", nil)
}

func getBasicData(ctx *gin.Context) {
	data, err := webservice.GetHomeData(ctx)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
		return
	}

	resp.Ok(ctx, "获取成功", data)
}

func redirectImgReq(ctx *gin.Context) {
	imgId := ctx.Param("img_id")

	preSignUrl, err := adminservices.GetImgPresignUrlById(ctx, imgId)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
	}

	resp.RedirectUrl(ctx, preSignUrl)
}

// getBlogData 获取博客详细数据
// @param ctx *gin.Context - Gin上下文
// @return 无返回值，通过resp包响应数据
func getBlogData(ctx *gin.Context) {
	// 从URL参数中获取博客ID
	blogId := ctx.Param("blog_id")

	// 调用service层获取博客数据和预签名URL
	blogData, preUrl, err := webservice.GetBlogDataById(ctx, blogId)
	if err != nil {
		// 如果获取失败，返回错误信息
		resp.Err(ctx, "获取失败", err.Error())
		return
	}

	// 获取成功，返回博客数据和预签名URL
	resp.Ok(ctx, "获取成功", map[string]any{
		"blog_data":    blogData,
		"pre_sign_url": preUrl,
	})
}

// searchContent 搜索内容
// RESTful API: POST /web/search/:content
//
// @param ctx *gin.Context - Gin上下文
// @return 无返回值，通过resp包响应搜索结果
func searchContent(ctx *gin.Context) {
	// 1. 获取搜索关键词（从URL参数）
	content := ctx.Param("content")
	if content == "" {
		resp.BadRequest(ctx, "搜索关键词不能为空", nil)
		return
	}

	// URL解码搜索关键词
	decodedContent, err := url.QueryUnescape(content)
	if err != nil {
		resp.BadRequest(ctx, "搜索关键词格式错误", err.Error())
		return
	}

	// 2. 构建搜索请求
	searchReq := searchengine.SearchRequest{
		Query:     decodedContent,
		Size:      0,                                // 0表示返回所有结果
		From:      0,                                // 从第一个开始
		Fields:    searchengine.DefaultSearchFields, // 返回全部字段
		Highlight: true,                             // 默认启用高亮
	}

	// 3. 执行搜索
	searchResult, err := searchengine.Search(searchReq)
	if err != nil {
		resp.Err(ctx, "搜索失败", err.Error())
		return
	}

	// 4. 转换搜索结果为map
	results := make([]map[string]any, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		result := map[string]any{
			"id":         hit.ID,
			"highlights": make(map[string][]string),
		}

		// 提取标题
		if title, exists := hit.Fields[searchengine.FieldTitle]; exists {
			if titleStr, ok := title.(string); ok {
				result["title"] = titleStr
			}
		}

		// 处理高亮片段
		if len(hit.Fragments) > 0 {
			highlights := make(map[string][]string)
			for field, fragments := range hit.Fragments {
				highlightList := make([]string, len(fragments))
				for i, fragment := range fragments {
					highlightList[i] = string(fragment)
				}
				highlights[field] = highlightList
			}
			result["highlights"] = highlights
		}

		results = append(results, result)
	}

	// 5. 构建响应map
	response := map[string]any{
		"results": results,
		"time_ms": searchResult.TimeMs,
	}

	// 6. 返回搜索结果
	resp.Ok(ctx, "搜索成功", response)
}
