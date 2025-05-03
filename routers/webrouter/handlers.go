package webrouter

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/internal/services/sysservices"
	"h2blog_server/internal/services/webservice"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
)

func getSysStatus(ctx *gin.Context) {
	if config.User.Username == "" {
		resp.Err(ctx, "服务状态异常，请检查配置文件", nil)
		return
	}

	resp.Ok(ctx, "获取成功", nil)
}

func getHomeData(ctx *gin.Context) {
	data, err := webservice.GetHomeData(ctx)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
		return
	}

	resp.Ok(ctx, "获取成功", data)
}

func redirectImgReq(ctx *gin.Context) {
	imgId := ctx.Param("img_id")

	url, err := sysservices.GetImgPresignUrlById(ctx, imgId)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
	}

	resp.RedirectUrl(ctx, url)
}
