package webrouter

import "github.com/gin-gonic/gin"

func Router(e *gin.Engine) {
	configGroup := e.Group("/web")

	// 获取用户基本信息
	configGroup.GET("/user-basic-info", userBasicInfo)
}
