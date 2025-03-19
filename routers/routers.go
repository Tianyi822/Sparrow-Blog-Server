package routers

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/env"
	"h2blog_server/pkg/middleware"
)

// Option 接受一个 *gin.Engine 类型的参数
type Option func(engine *gin.Engine)

var options []Option

// IncludeOpts 接受多个 Option 类型的参数
func IncludeOpts(opts ...Option) {
	// 将传入的 Option 参数追加到全局变量 options 中
	options = append(options, opts...)
}

func InitRouter() *gin.Engine {
	// 创建一个没有任何中间件的路由
	r := gin.New()

	// 添加自定义的中间件
	switch env.CurrentEnv {
	case env.RuntimeEnv:
		r.Use(middleware.Logger(), middleware.RunTimeCors(), gin.Recovery())
	case env.ConfigServerEnv:
		r.Use(middleware.ConfigServiceCors(), gin.Recovery())
	}

	for _, opt := range options {
		opt(r)
	}
	// 注册完成后重置 options
	options = make([]Option, 0)
	return r
}
