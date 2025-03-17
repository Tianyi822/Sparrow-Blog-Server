package webRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	webGroup := e.Group("/web")

	webGroup.GET("/status", status)
}
