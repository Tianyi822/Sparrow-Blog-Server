package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"h2blog_server/pkg/config"
)

// InitiatedStepCors 初始化阶段的CORS中间件
// 允许所有来源访问，仅支持GET和POST方法
// 允许Origin和Content-Type请求头
func InitiatedStepCors() gin.HandlerFunc {
	c := cors.Config{
		AllowAllOrigins: true, // 允许所有来源
		AllowMethods:    []string{"GET", "POST"},
		AllowHeaders:    []string{"Origin", "Content-Type"},
	}

	return cors.New(c)
}

// RunTimeCors 运行时的CORS中间件
// 根据配置文件设置允许的:
// - 来源域名
// - HTTP方法
// - 请求头
func RunTimeCors() gin.HandlerFunc {
	c := cors.Config{
		AllowOrigins: config.Server.Cors.Origins,
		AllowMethods: config.Server.Cors.Methods,
		AllowHeaders: config.Server.Cors.Headers,
	}

	return cors.New(c)
}
