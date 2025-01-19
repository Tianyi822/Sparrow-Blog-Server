package img_routers

import (
	"github.com/gin-gonic/gin"
	"h2blog/internal/services/imgService"
	"h2blog/pkg/resp"
	"h2blog/routers/tools"
)

func uploadImages(ctx *gin.Context) {
	// 从 RawData 中获取到图片信息
	imgsDto, err := tools.GetImgsDto(ctx)
	if err != nil {
		return
	}

	// 压缩图片并保存
	imgVo, err := imgService.ConvertAndAddImg(ctx, imgsDto)
	if err != nil {
		resp.Err(ctx, err.Error(), nil)
		return
	}

	resp.Ok(ctx, "上传成功", imgVo)
}
