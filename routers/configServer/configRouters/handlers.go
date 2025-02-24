package configRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/configServer/configAnalyze"
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
	}
	config.Server.Port = port

	// 解析 Token 密钥
	if err = configAnalyze.AnalyzeTokenKey(tokenKey); err != nil {
		resp.BadRequest(ctx, "Token 密钥配置错误", err)
	}
	config.Server.TokenKey = tokenKey

	dur, err := configAnalyze.AnalyzeTokenExpireDuration(tokenExpireDuration)
	if err != nil {
		resp.BadRequest(ctx, "Token 过期时间配置错误", err)
	}
	config.Server.TokenExpireDuration = dur

	if err = configAnalyze.AnalyzeCorsOrigins(corsOrigins); err != nil {
		resp.BadRequest(ctx, "跨域源配置错误", err)
	}
	config.Server.Cors.Origins = corsOrigins

	config.Server.Cors.Headers = []string{"Content-Type", "Authorization", "X-CSRF-Token"}
	config.Server.Cors.Methods = []string{"POST", "PUT", "DELETE", "GET"}

	resp.Ok(ctx, "配置完成", config.Server)
}

// closeConfigServer 关闭配置服务
func closeConfigServer(ctx *gin.Context) {
	resp.Ok(ctx, "关闭配置服务", nil)
	// 配置完成，通知关闭配置服务
	env.CompletedConfigSign <- true
}
