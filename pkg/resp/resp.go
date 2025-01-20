package resp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ResponseMsg
// @desc 响应消息体
type ResponseMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func MakeResp(ctx *gin.Context, statusCode int, msg string, data any) {
	ctx.JSON(
		statusCode,
		&ResponseMsg{
			Code: statusCode,
			Msg:  msg,
			Data: data,
		},
	)

	ctx.Abort()
}

func Ok(ctx *gin.Context, msg string, data any) {
	MakeResp(ctx, http.StatusOK, msg, data)
}

func Err(ctx *gin.Context, msg string, data any) {
	MakeResp(ctx, http.StatusInternalServerError, msg, data)
}

func BadRequest(ctx *gin.Context, msg string, data any) {
	MakeResp(ctx, http.StatusBadRequest, msg, data)
}

// TokenIsUnauthorized Token 未验证通过
func TokenIsUnauthorized(ctx *gin.Context, msg string, data any) {
	MakeResp(ctx, http.StatusUnauthorized, msg, data)
}

// RedirectUrl 重定向
func RedirectUrl(ctx *gin.Context, url string) {
	ctx.Redirect(http.StatusFound, url)
}
