package email

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/utils"
	"h2blog_server/storage"
	"time"

	"gopkg.in/gomail.v2"
)

func SendVerificationCodeEmail(ctx context.Context, email string) error {
	code := utils.GenRandomString(6)

	switch env.CurrentEnv {
	case env.ConfigServerEnv:
		if env.VerificationCode == "" {
			env.VerificationCode = code
		} else {
			code = env.VerificationCode
		}
	case env.RuntimeEnv:
		c, err := storage.Storage.Cache.GetString(ctx, "config-server-verification-code")
		if err != nil {
			err = storage.Storage.Cache.SetWithExpired(ctx, "config-server-verification-code", code, 5*time.Minute)
			if err != nil {
				msg := fmt.Sprintf("缓存验证码失败: %v", err)
				return errors.New(msg)
			}
		}
		code = c
	}

	// 创建邮件内容
	m := gomail.NewMessage()
	// 发件人
	m.SetHeader("From", config.User.SysEmailAccount)
	// 收件人
	m.SetHeader("To", email)
	// 主题
	m.SetHeader("Subject", "博客验证码")
	// HTML 正文，包含验证码，验证码加粗并使用彩色显示
	htmlContent := fmt.Sprintf("<h1>您的验证码为：<span style=\"color: red;\">%s</span></h1>", code)
	m.SetBody("text/html", htmlContent)
	// 配置 SMTP 服务器
	d := gomail.NewDialer(config.User.SysEmailSmtp, config.User.SysEmailPort, config.User.SysEmailAccount, config.User.SysEmailAuthCode)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
