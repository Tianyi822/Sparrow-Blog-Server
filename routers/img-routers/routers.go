package img_routers

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	// 添加博客路由
	imgGroup := e.Group("/img")

	// 图片上传
	imgGroup.POST("/upload", uploadImages)
}
