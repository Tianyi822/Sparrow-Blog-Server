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

		galleryGroup.GET("/all-imgs", getAllImgs)

		galleryGroup.DELETE("/:img_id", deleteImg)

		galleryGroup.PUT("/:img_id", renameImg)
	}
}
