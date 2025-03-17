package emailRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {

	// 邮箱接口
	emailGroup := e.Group("/email")

	// 发送验证码
	emailGroup.POST("/verification-code", sendVerificationCode)
}
