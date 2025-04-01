package imgrouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/internal/services/imgservices"
	"h2blog_server/pkg/resp"
)

func redirectImgReq(ctx *gin.Context) {
	imgId := ctx.Param("img_id")

	url, err := imgservices.GetPresignUrlById(ctx, imgId)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
	}

	resp.RedirectUrl(ctx, url)
}
