package imgrouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {

	// 邮箱接口
	imgGroup := e.Group("/img")

	// 发送验证码
	imgGroup.GET("/get/:img_id", redirectImgReq)
}
