package adminrouter

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/routers/middleware"
)

func Routers(e *gin.Engine) {
	adminGroup := e.Group("/admin")

	{
		loginGroup := adminGroup.Group("/login")

		loginGroup.POST("", login)

		loginGroup.GET("/user-info", getUserInfo)

		loginGroup.POST("/verification-code", sendLoginVerificationCode)

		// 退出登录需要鉴权 token
		adminGroup.Use(middleware.AnalyzeJWT()).GET("/logout", logout)
	}

	{
		ossGroup := adminGroup.Group("/oss")

		ossGroup.Use(middleware.AnalyzeJWT())

		ossGroup.GET("/pre_sign_url/:file_name/type/:file_type", genPresignPutUrl)
	}

	{
		postsGroup := adminGroup.Group("/posts")

		postsGroup.GET("/all-blogs", getAllBlogs)

		postsGroup.GET("/change-blog-state/:blog_id", changeBlogState)

		postsGroup.GET("/set-top/:blog_id", setTop)

		postsGroup.DELETE("/delete/:blog_id", deleteBlog)
	}

	{
		editGroup := adminGroup.Group("/edit")

		editGroup.GET("/all-tags-categories", getAllTagsCategories)

		editGroup.POST("/update-or-add-blog", updateOrAddBlog)

		editGroup.GET("/blog-data/:blog_id", getBlogData)
	}

	{
		galleryGroup := adminGroup.Group("/gallery")

		galleryGroup.POST("/add", addImgs)

		galleryGroup.GET("/all-imgs", getAllImgs)

		galleryGroup.DELETE("/:img_id", deleteImg)

		galleryGroup.PUT("/:img_id", renameImg)

		galleryGroup.GET("/is-exist/:img_name", isExist)
	}

	{
		settingGroup := adminGroup.Group("/setting")

		settingGroup.GET("/user/config", getUserConfig)

		settingGroup.PUT("/user/config", updateUserConfig)

		settingGroup.POST("/user/verify-new-email", verifyNewEmail)

		settingGroup.PUT("/user/visual", updateUserVisuals)

		settingGroup.GET("/server/config", getServerConfig)

		settingGroup.PUT("/server/config", updateServerConfig)

		settingGroup.POST("/user/verify-new-smtp-config", verifyNewSmtpConfig)

		settingGroup.GET("/logger/config", getLoggerConfig)

		settingGroup.PUT("/logger/config", updateLoggerConfig)

		settingGroup.GET("/mysql/config", getMysqlConfig)

		settingGroup.PUT("/mysql/config", updateMysqlConfig)

		settingGroup.GET("/oss/config", getOssConfig)

		settingGroup.PUT("/oss/config", updateOssConfig)

		settingGroup.GET("/cache/config", getCacheConfig)

		settingGroup.PUT("/cache/config", updateCacheConfig)
	}
}
