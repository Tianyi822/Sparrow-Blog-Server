package img_routers

import (
	"github.com/gin-gonic/gin"
	"h2blog/pkg/resp"
	"h2blog/routers/tools"
)

func uploadImages(ctx *gin.Context) {
	// Get imgs data from raw data
	imgsDto, err := tools.GetImgsDto(ctx)
	if err != nil {
		return
	}

	resp.Ok(ctx, "上传成功", imgsDto)
}
