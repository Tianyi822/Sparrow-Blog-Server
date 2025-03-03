package email

import (
	"h2blog_server/email"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/tools"

	"github.com/gin-gonic/gin"
)

func sendVerificationCode(ctx *gin.Context) {
	data, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		return
	}

	// 发送验证码到邮箱中
	if err = email.SendVerificationCodeEmail(ctx, data["email"].(string)); err != nil {
		resp.Err(ctx, "发送验证码失败", err)
		return
	}

	resp.Ok(ctx, "已发送验证码", nil)
}
