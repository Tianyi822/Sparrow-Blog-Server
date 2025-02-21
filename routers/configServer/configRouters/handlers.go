package configRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/env"
	"h2blog_server/pkg/resp"
)

func configBase(ctx *gin.Context) {
	resp.Ok(ctx, "配置完成", nil)
}

// closeConfigServer 关闭配置服务
func closeConfigServer(ctx *gin.Context) {
	resp.Ok(ctx, "关闭配置服务", nil)
	// 配置完成，通知关闭配置服务
	env.CompletedConfigSign <- true
}
