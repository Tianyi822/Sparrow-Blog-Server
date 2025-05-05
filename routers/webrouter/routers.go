package webrouter

import "github.com/gin-gonic/gin"

func Router(e *gin.Engine) {
	webGroup := e.Group("/web")

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

	webGroup.GET("/basic-data", getBasicData)
}
