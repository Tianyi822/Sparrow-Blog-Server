package webrouter

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/internal/services/adminservices"
	"h2blog_server/internal/services/webservice"
	"h2blog_server/pkg/config"
	"h2blog_server/routers/resp"
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

	url, err := adminservices.GetImgPresignUrlById(ctx, imgId)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
	}

	resp.RedirectUrl(ctx, url)
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
