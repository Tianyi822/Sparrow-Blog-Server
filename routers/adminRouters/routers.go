package adminRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	adminGroup := e.Group("/admin")

	{
		loginGroup := adminGroup.Group("/login")

		loginGroup.POST("/verification-code", sendVerificationCode)

		loginGroup.POST("/login", login)
	}

	{
		postsGroup := adminGroup.Group("/posts")

		postsGroup.GET("/all-blogs", getAllBlogs)
	}
}
