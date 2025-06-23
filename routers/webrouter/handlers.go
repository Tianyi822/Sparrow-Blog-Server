package webrouter

import (
	"net/url"

	"sparrow_blog_server/internal/model/vo"
	"sparrow_blog_server/internal/services/adminservices"
	"sparrow_blog_server/internal/services/webservice"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/email"
	"sparrow_blog_server/routers/resp"
	"sparrow_blog_server/routers/tools"
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

		// 提取文章封面图片 ID
		if imgId, exists := hit.Fields[searchengine.FieldImgId]; exists {
			if imgIdStr, ok := imgId.(string); ok {
				result["img_id"] = imgIdStr
			}
		}

		// 处理高亮片段
		if len(hit.Fragments) > 0 {
			highlights := make(map[string][]string)
			for field, fragments := range hit.Fragments {
				highlightList := make([]string, len(fragments))
				copy(highlightList, fragments)
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

// getAllDisplayedFriendLinks 获取所有显示状态为 true 的友链
// @param ctx *gin.Context - Gin上下文
// @return 无返回值，通过resp包响应友链数据
func getAllDisplayedFriendLinks(ctx *gin.Context) {
	// 调用webservice层获取显示状态为 true 的友链数据
	friendLinkDtos, err := webservice.GetDisplayedFriendLinks(ctx)
	if err != nil {
		resp.Err(ctx, "获取友链失败", err.Error())
		return
	}

	// 将DTO列表转换为VO列表，以便前端使用
	friendLinkVos := make([]vo.FriendLinkVo, 0, len(friendLinkDtos))
	for _, friendLinkDto := range friendLinkDtos {
		friendLinkVo := vo.FriendLinkVo{
			FriendLinkId:    friendLinkDto.FriendLinkId,
			FriendLinkName:  friendLinkDto.FriendLinkName,
			FriendLinkUrl:   friendLinkDto.FriendLinkUrl,
			FriendAvatarUrl: friendLinkDto.FriendAvatarUrl,
			FriendDescribe:  friendLinkDto.FriendDescribe,
			Display:         friendLinkDto.Display,
		}
		friendLinkVos = append(friendLinkVos, friendLinkVo)
	}

	// 返回成功响应
	resp.Ok(ctx, "获取友链成功", friendLinkVos)
}

// applyFriendLink 申请友链
// @param ctx *gin.Context - Gin上下文
// @return 无返回值，通过resp包响应结果
func applyFriendLink(ctx *gin.Context) {
	// 使用tools包中的GetFriendLinkDto方法获取友链DTO
	friendLinkDto, err := tools.GetFriendLinkDto(ctx)
	if err != nil {
		// GetFriendLinkDto内部已经处理了错误响应，这里直接返回
		return
	}

	// 调用service层处理友链申请
	if err := webservice.ApplyFriendLink(ctx, friendLinkDto); err != nil {
		resp.Err(ctx, "友链申请失败: "+err.Error(), nil)
		return
	}

	// 异步发送邮件通知管理员
	go func() {
		emailData := email.FriendLinkData{
			Name:        friendLinkDto.FriendLinkName,
			URL:         friendLinkDto.FriendLinkUrl,
			AvatarURL:   friendLinkDto.FriendAvatarUrl,
			Description: friendLinkDto.FriendDescribe,
		}

		// 发送邮件通知
		if err := email.SendFriendLinkNotificationBySys(ctx.Copy(), emailData); err != nil {
			// 邮件发送失败只记录日志，不影响主流程
			// 这里可以添加日志记录
			_ = err
		}
	}()

	// 返回成功响应
	resp.Ok(ctx, "友链申请成功，请等待管理员审核", nil)
}
