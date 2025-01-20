package imgRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog/internal/services/imgService"
	"h2blog/pkg/resp"
	"h2blog/routers/tools"
)

// uploadImages 上传图片
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

func deleteImgs(ctx *gin.Context) {
	// 从 RawData 中获取到数据
	data, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		return
	}

	// 修改类型断言方式
	idsInterface, ok := data["ids"].([]any)
	if !ok {
		resp.Err(ctx, "ids 格式错误", nil)
		return
	}

	// 手动转换为[]string
	ids := make([]string, len(idsInterface))
	for i, v := range idsInterface {
		if str, ok := v.(string); ok {
			ids[i] = str
		} else {
			resp.Err(ctx, "ids 包含非字符串元素", nil)
			return
		}
	}

	// 根据 id 批量删除数据
	imgInfosVo, err := imgService.DeleteImgs(ctx, ids)
	if err != nil {
		resp.Err(ctx, err.Error(), nil)
		return
	}

	resp.Ok(ctx, "删除成功", imgInfosVo)
}

func renameImgName(ctx *gin.Context) {
	// 从 RawData 中获取到数据
	dto, err := tools.GetImgDto(ctx)
	if err != nil {
		return
	}

	if len(dto.ImgName) == 0 {
		resp.BadRequest(ctx, "图片名称不能为空", nil)
		return
	}

	imgInfoVo, err := imgService.RenameImgs(ctx, dto.ImgId, dto.ImgName)
	if err != nil {
		resp.Err(ctx, err.Error(), nil)
		return
	}

	resp.Ok(ctx, "修改成功", imgInfoVo)
}
