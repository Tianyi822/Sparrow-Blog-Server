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
	}
}
