package blogRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	// 添加博客路由
	blogGroup := e.Group("/blog")

	blogGroup.GET("/:blog_id", getBlogById)
}
