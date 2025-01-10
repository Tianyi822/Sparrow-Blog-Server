package resp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
	200 - 成功
	500 - 服务器错误
	400 - 访问错误
	409 - 重复访问错误
	401 - Token 未验证通过
*/

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

func ConflictRequest(ctx *gin.Context, msg string, data any) {
	MakeResp(ctx, http.StatusConflict, msg, data)
}

// TokenIsUnauthorized Token 未验证通过
func TokenIsUnauthorized(ctx *gin.Context, msg string, data any) {
	MakeResp(ctx, http.StatusUnauthorized, msg, data)
}
