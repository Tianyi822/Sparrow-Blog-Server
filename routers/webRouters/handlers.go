package webRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/env"
	"h2blog_server/pkg/resp"
)

// status 处理状态请求
// 该函数通过 HTTP 上下文 ctx 返回当前环境状态
func status(ctx *gin.Context) {
	resp.Ok(ctx, env.CurrentEnv, nil)
}
