package configRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/configServer/configAnalyze"
	"h2blog_server/routers/tools"
	"strings"
)

func configBase(ctx *gin.Context) {
	portStr := strings.TrimSpace(ctx.PostForm("server.port"))
	tokenKey := strings.TrimSpace(ctx.PostForm("server.token_key"))
	tokenExpireDuration := strings.TrimSpace(ctx.PostForm("server.token_expire_duration"))
	corsOrigins := ctx.PostFormArray("server.cors.origins")

	// 解析端口
	port, err := configAnalyze.AnalyzePort(portStr)
	if err != nil {
		resp.BadRequest(ctx, "端口配置错误", err)
		return
	}
	config.Server.Port = port

	// 解析 Token 密钥
	if err = configAnalyze.AnalyzeTokenKey(tokenKey); err != nil {
		resp.BadRequest(ctx, "Token 密钥配置错误", err)
		return
	}
	config.Server.TokenKey = tokenKey

	dur, err := configAnalyze.AnalyzeTokenExpireDuration(tokenExpireDuration)
	if err != nil {
		resp.BadRequest(ctx, "Token 过期时间配置错误", err)
		return
	}
	config.Server.TokenExpireDuration = dur

	if err = configAnalyze.AnalyzeCorsOrigins(corsOrigins); err != nil {
		resp.BadRequest(ctx, "跨域源配置错误", err)
		return
	}
	config.Server.Cors.Origins = corsOrigins

	config.Server.Cors.Headers = []string{"Content-Type", "Authorization", "X-CSRF-Token"}
	config.Server.Cors.Methods = []string{"POST", "PUT", "DELETE", "GET"}

	resp.Ok(ctx, "配置完成", config.Server)
}

func configUser(ctx *gin.Context) {
	config.User.Username = strings.TrimSpace(ctx.PostForm("user.username"))

	email := strings.TrimSpace(ctx.PostForm("user.email"))
	if err := configAnalyze.AnalyzeEmail(email); err != nil {
		resp.BadRequest(ctx, "邮箱配置错误", err)
		return
	}
	config.User.Email = email

	resp.Ok(ctx, "配置完成", config.User)
}

func configOss(ctx *gin.Context) {
	// OSS 基础配置
	config.Oss.Endpoint = strings.TrimSpace(ctx.PostForm("oss.endpoint"))
	config.Oss.Region = strings.TrimSpace(ctx.PostForm("oss.region"))
	config.Oss.AccessKeyId = strings.TrimSpace(ctx.PostForm("oss.access_key_id"))
	config.Oss.AccessKeySecret = strings.TrimSpace(ctx.PostForm("oss.access_key_secret"))
	config.Oss.Bucket = strings.TrimSpace(ctx.PostForm("oss.bucket"))
	if err := configAnalyze.AnalyzeOssConfig(config.Oss); err != nil {
		resp.BadRequest(ctx, "OSS 配置错误", err)
		return
	}

	// OSS 路径配置
	imageOssPath := strings.TrimSpace(ctx.PostForm("oss.image_oss_path"))
	if err := configAnalyze.AnalyzeOssPath(imageOssPath); err != nil {
		resp.BadRequest(ctx, "图片 OSS 路径配置错误", err)
		return
	}
	config.Oss.ImageOssPath = imageOssPath

	blogOssPath := strings.TrimSpace(ctx.PostForm("oss.blog_oss_path"))
	if err := configAnalyze.AnalyzeOssPath(blogOssPath); err != nil {
		resp.BadRequest(ctx, "博客 OSS 路径配置错误", err)
		return
	}
	config.Oss.BlogOssPath = blogOssPath

	// OSS 下的 webp 文件配置
	webpEnable, err := tools.GetIntFromPostForm(ctx, "oss.webp.enable")
	if err != nil {
		resp.BadRequest(ctx, "WebP 启用配置错误", err)
		return
	}
	config.Oss.WebP.Enable = webpEnable == 1

	webpQuality, err := tools.GetFloatFromPostForm(ctx, "oss.webp.quality")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩质量配置错误", err)
		return
	}
	config.Oss.WebP.Quality = webpQuality

	webpSize, err := tools.GetFloatFromPostForm(ctx, "oss.webp.size")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩后大小配置错误", err)
		return
	}
	config.Oss.WebP.Size = webpSize

	resp.Ok(ctx, "配置完成", config.Oss)
}

// closeConfigServer 关闭配置服务
func closeConfigServer(ctx *gin.Context) {
	resp.Ok(ctx, "关闭配置服务", nil)
	// 配置完成，通知关闭配置服务
	env.CompletedConfigSign <- true
}
