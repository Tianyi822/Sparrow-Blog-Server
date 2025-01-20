package imgRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	// 添加博客路由
	imgGroup := e.Group("/img")

	// 图片上传
	imgGroup.POST("/upload", uploadImages)
	// 图片删除
	imgGroup.DELETE("/delete", deleteImgs)
	// 图片重命名
	imgGroup.PUT("/rename", renameImgName)
	// 图片模糊查询
	imgGroup.GET("/get_list/:name", getImgsByNameLike)
	// 根据 ID 获取图片
	imgGroup.GET("/get/:id", getImgByID)
}
