package adminRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	adminGroup := e.Group("/admin")

	adminGroup.POST("/verification-code", sendVerificationCode)

	adminGroup.POST("/login", login)

	adminGroup.GET("/all-blogs", getAllBlogs)
}
