package webrouter

import "github.com/gin-gonic/gin"

func Router(e *gin.Engine) {
	webGroup := e.Group("/web")

	webGroup.GET("/basic-data", getBasicData)

	{
		sysGroup := webGroup.Group("/sys")

		sysGroup.GET("/status", getSysStatus)
	}

	{
		imageGroup := webGroup.Group("/img")

		imageGroup.GET("/get/:img_id", redirectImgReq)
	}

	{
		blogGroup := webGroup.Group("/blog")

		blogGroup.GET("/:blog_id", getBlogData)
	}

	{
		searchGroup := webGroup.Group("/search")

		searchGroup.GET("/:content", searchContent)
	}

	{
		friendLinkGroup := webGroup.Group("/friend-link")

		friendLinkGroup.GET("/all", getAllDisplayedFriendLinks)

		friendLinkGroup.POST("/apply", applyFriendLink)
	}

	{
		commentGroup := webGroup.Group("/comment")

		// 根据博客ID获取所有评论及子评论
		commentGroup.GET("/:blog_id", getCommentsByBlogId)

		// 添加评论
		commentGroup.POST("", addComment)

		// 回复评论
		commentGroup.POST("/reply", replyComment)
	}
}
