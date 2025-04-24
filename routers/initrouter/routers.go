package initrouter

import (
	"github.com/gin-gonic/gin"
)

func Routers(e *gin.Engine) {
	initGroup := e.Group("/init")

	initGroup.POST("/server", initServer)

	initGroup.POST("/send-code", sendCode)

	initGroup.POST("/user", initUser)

	initGroup.POST("/mysql", initMysql)

	initGroup.POST("/oss", initOss)

	initGroup.POST("/cache", initCache)

	initGroup.POST("/logger", initLogger)

	initGroup.GET("/complete-config", completeInit)
}
