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
	serverConfig := &config.ServerConfigData{}

	// 以下两项不需要前端传入配置，直接写死
	serverConfig.Cors.Headers = []string{"Content-Type", "Authorization", "X-CSRF-Token"}
	serverConfig.Cors.Methods = []string{"POST", "PUT", "DELETE", "GET"}

	// 解析端口
	portStr := strings.TrimSpace(ctx.PostForm("server.port"))
	port, err := configAnalyze.AnalyzePort(portStr)
	if err != nil {
		resp.BadRequest(ctx, "端口配置错误", err)
		return
	}
	serverConfig.Port = port

	// 解析 Token 密钥
	tokenKey := strings.TrimSpace(ctx.PostForm("server.token_key"))
	if err = configAnalyze.AnalyzeTokenKey(tokenKey); err != nil {
		resp.BadRequest(ctx, "Token 密钥配置错误", err)
		return
	}
	serverConfig.TokenKey = tokenKey

	tokenExpireDuration := strings.TrimSpace(ctx.PostForm("server.token_expire_duration"))
	dur, err := configAnalyze.AnalyzeTokenExpireDuration(tokenExpireDuration)
	if err != nil {
		resp.BadRequest(ctx, "Token 过期时间配置错误", err)
		return
	}
	serverConfig.TokenExpireDuration = dur

	corsOrigins := ctx.PostFormArray("server.cors.origins")
	if err = configAnalyze.AnalyzeCorsOrigins(corsOrigins); err != nil {
		resp.BadRequest(ctx, "跨域源配置错误", err.Error())
		return
	}
	serverConfig.Cors.Origins = corsOrigins

	// 完成配置，将配置添加到全局
	config.Server = serverConfig

	resp.Ok(ctx, "配置完成", config.Server)
}

func configUser(ctx *gin.Context) {
	userConfig := &config.UserConfigData{}

	userConfig.Username = strings.TrimSpace(ctx.PostForm("user.username"))

	email := strings.TrimSpace(ctx.PostForm("user.email"))
	if err := configAnalyze.AnalyzeEmail(email); err != nil {
		resp.BadRequest(ctx, "邮箱配置错误", err)
		return
	}
	userConfig.Email = email

	// 完成配置，将配置添加到全局
	config.User = userConfig

	resp.Ok(ctx, "配置完成", config.User)
}

func configOss(ctx *gin.Context) {
	ossConfig := &config.OssConfig{}

	// OSS 基础配置
	ossConfig.Endpoint = strings.TrimSpace(ctx.PostForm("oss.endpoint"))
	ossConfig.Region = strings.TrimSpace(ctx.PostForm("oss.region"))
	ossConfig.AccessKeyId = strings.TrimSpace(ctx.PostForm("oss.access_key_id"))
	ossConfig.AccessKeySecret = strings.TrimSpace(ctx.PostForm("oss.access_key_secret"))
	ossConfig.Bucket = strings.TrimSpace(ctx.PostForm("oss.bucket"))
	if err := configAnalyze.AnalyzeOssConfig(ossConfig); err != nil {
		resp.BadRequest(ctx, "OSS 配置错误", err)
		return
	}

	// OSS 路径配置
	imageOssPath := strings.TrimSpace(ctx.PostForm("oss.image_oss_path"))
	if err := configAnalyze.AnalyzeOssPath(imageOssPath); err != nil {
		resp.BadRequest(ctx, "图片 OSS 路径配置错误", err)
		return
	}
	ossConfig.ImageOssPath = imageOssPath

	blogOssPath := strings.TrimSpace(ctx.PostForm("oss.blog_oss_path"))
	if err := configAnalyze.AnalyzeOssPath(blogOssPath); err != nil {
		resp.BadRequest(ctx, "博客 OSS 路径配置错误", err)
		return
	}
	ossConfig.BlogOssPath = blogOssPath

	// OSS 下的 webp 文件配置
	webpEnable, err := tools.GetIntFromPostForm(ctx, "oss.webp.enable")
	if err != nil {
		resp.BadRequest(ctx, "WebP 启用配置错误", err)
		return
	}
	ossConfig.WebP.Enable = webpEnable == 1

	webpQuality, err := tools.GetFloatFromPostForm(ctx, "oss.webp.quality")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩质量配置错误", err)
		return
	}
	ossConfig.WebP.Quality = webpQuality

	webpSize, err := tools.GetFloatFromPostForm(ctx, "oss.webp.size")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩后大小配置错误", err)
		return
	}
	ossConfig.WebP.Size = webpSize

	// 完成配置，并将配置添加到全局
	config.Oss = ossConfig

	resp.Ok(ctx, "配置完成", config.Oss)
}

// closeConfigServer 关闭配置服务
func closeConfigServer(ctx *gin.Context) {
	resp.Ok(ctx, "关闭配置服务", nil)
	// 配置完成，通知关闭配置服务
	env.CompletedConfigSign <- true
}
