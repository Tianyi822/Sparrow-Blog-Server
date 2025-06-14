package adminrouter

import (
	"github.com/gin-gonic/gin"
	"sparrow_blog_server/env"
	"sparrow_blog_server/routers/middleware"
)

func Routers(e *gin.Engine) {
	adminGroup := e.Group("/admin")

	{
		loginGroup := adminGroup.Group("/login")

		loginGroup.POST("", login)

		loginGroup.GET("/user-info", getUserInfo)

		loginGroup.POST("/verification-code", sendLoginVerificationCode)
	}

	{
		logoutGroup := adminGroup.Group("/logout")

		if env.CurrentEnv == env.ProdEnv {
			logoutGroup.Use(middleware.AnalyzeJWT())
		}

		logoutGroup.GET("", logout)
	}

	{
		ossGroup := adminGroup.Group("/oss")

		if env.CurrentEnv == env.ProdEnv {
			ossGroup.Use(middleware.AnalyzeJWT())
		}

		ossGroup.GET("/pre_sign_url/:file_name/type/:file_type", genPresignPutUrl)
	}

	{
		postsGroup := adminGroup.Group("/posts")

		if env.CurrentEnv == env.ProdEnv {
			postsGroup.Use(middleware.AnalyzeJWT())
		}

		postsGroup.GET("/all-blogs", getAllBlogs)

		postsGroup.GET("/change-blog-state/:blog_id", changeBlogState)

		postsGroup.GET("/set-top/:blog_id", setTop)

		postsGroup.DELETE("/delete/:blog_id", deleteBlog)
	}

	{
		editGroup := adminGroup.Group("/edit")

		if env.CurrentEnv == env.ProdEnv {
			editGroup.Use(middleware.AnalyzeJWT())
		}

		editGroup.GET("/all-tags-categories", getAllTagsCategories)

		editGroup.POST("/update-or-add-blog", updateOrAddBlog)

		editGroup.GET("/blog-data/:blog_id", getBlogData)
	}

	{
		galleryGroup := adminGroup.Group("/gallery")

		if env.CurrentEnv == env.ProdEnv {
			galleryGroup.Use(middleware.AnalyzeJWT())
		}

		galleryGroup.POST("/add", addImgs)

		galleryGroup.GET("/all-imgs", getAllImgs)

		galleryGroup.DELETE("/:img_id", deleteImg)

		galleryGroup.PUT("/:img_id", renameImg)

		galleryGroup.GET("/is-exist/:img_name", isExist)
	}

	{
		settingGroup := adminGroup.Group("/setting")

		if env.CurrentEnv == env.ProdEnv {
			settingGroup.Use(middleware.AnalyzeJWT())
		}

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

		settingGroup.GET("/cache-index/config", getCacheAndIndexConfig)

		settingGroup.PUT("/cache-index/config", updateCacheAndIndexConfig)
	}
}
