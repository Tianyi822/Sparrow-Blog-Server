package blog_routers

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	// 添加博客路由
	blogGroup := e.Group("/blog")

	blogGroup.GET("/:blog_id", getBlogById)

	blogGroup.PUT("/modify", modifyBlog)

	blogGroup.DELETE("/:blog_id", deleteBlogById)

	blogGroup.POST("/add", addBlogInfo)
}
