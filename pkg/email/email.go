package email

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/utils"
	"sparrow_blog_server/storage"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

// SendVerificationCodeByArgs 发送包含验证码的电子邮件。
// 参数说明：
//   - ctx: 上下文对象，用于控制请求的生命周期。
//   - email: 收件人的电子邮件地址。
//   - smtpAccount: SMTP服务器的账户名。
//   - smtpAddress: SMTP服务器的地址。
//   - smtpAuthCode: SMTP服务器的授权码。
//   - smtpPort: SMTP服务器的端口号。
//
// 返回值：
//   - error: 如果发送邮件过程中发生错误，则返回错误信息；否则返回nil。
func SendVerificationCodeByArgs(ctx context.Context, email, smtpAccount, smtpAddress, smtpAuthCode string, smtpPort uint16) error {
	// 生成一个长度为20的随机验证码，基于用户邮箱和当前时间。
	code, err := utils.HashWithLength(config.User.UserEmail+time.Now().String(), 20)
	if err != nil {
		return err
	}

	// 如果缓存中不存在验证码，则将生成的验证码存储到缓存中，并设置5分钟的过期时间。
	c, getErr := storage.Storage.Cache.GetString(ctx, storage.VerificationCodeKey)
	if getErr != nil {
		setErr := storage.Storage.Cache.SetWithExpired(ctx, storage.VerificationCodeKey, code, 5*time.Minute)
		if setErr != nil {
			msg := fmt.Sprintf("缓存验证码失败: %v", err)
			return errors.New(msg)
		}
	} else {
		code = c
	}

	// 解析HTML模板，准备渲染验证码。
	tmpl, err := template.New("email").Parse(VerificationCodeTemplate)
	if err != nil {
		return err
	}

	// 创建一个字符串构建器，用于存储渲染后的HTML内容。
	var htmlContent strings.Builder

	// 执行模板渲染，将验证码插入到HTML模板中。
	err = tmpl.Execute(&htmlContent, struct {
		Code string
	}{Code: template.HTMLEscapeString(code)})
	if err != nil {
		return err
	}

	// 调用SendContent函数发送包含验证码的邮件。
	return sendContent(email, htmlContent.String(), VerificationCodeSubject, smtpAccount, smtpAddress, smtpAuthCode, smtpPort)
}

// sendContent 发送邮件内容到指定邮箱。
// 参数说明：
//   - email: 收件人的邮箱地址。
//   - content: 邮件的正文内容，支持 HTML 格式。
//   - subject: 邮件主题。
//   - smtpAccount: SMTP 服务器的发件人账号（通常是邮箱地址）。
//   - smtpAddress: SMTP 服务器的地址（如 smtp.example.com）。
//   - smtpAuthCode: SMTP 服务器的授权码或密码。
//   - smtpPort: SMTP 服务器的端口号（如 465 或 587）。
//
// 返回值：
//   - error: 如果发送邮件失败，则返回错误信息；否则返回 nil。
func sendContent(email, content, subject, smtpAccount, smtpAddress, smtpAuthCode string, smtpPort uint16) error {
	// 创建邮件内容
	m := gomail.NewMessage()

	// 设置邮件头部信息，包括发件人、收件人和主题
	m.SetHeader("From", smtpAccount)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)

	// 设置邮件正文为 HTML 格式
	m.SetBody("text/html", content)

	// 配置 SMTP 服务器连接信息
	d := gomail.NewDialer(smtpAddress, int(smtpPort), smtpAccount, smtpAuthCode)

	// 尝试连接 SMTP 服务器并发送邮件
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

// SendVerificationCodeBySys 发送验证码邮件给指定的邮箱地址。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据；
//   - email: 目标邮箱地址，验证码将发送到该邮箱；
//
// 返回值:
//   - error: 如果发送过程中出现错误，则返回具体的错误信息；否则返回 nil。
func SendVerificationCodeBySys(ctx context.Context) error {
	// 调用 SendVerificationCodeByArgs 函数发送验证码邮件，
	// 使用系统配置中的 SMTP 账号、地址、授权码和端口信息。
	if err := SendVerificationCodeByArgs(
		ctx,
		config.User.UserEmail,
		config.Server.SmtpAccount,
		config.Server.SmtpAddress,
		config.Server.SmtpAuthCode,
		config.Server.SmtpPort,
	); err != nil {
		return err
	}

	return nil
}

// FriendLinkData 友链信息结构体
type FriendLinkData struct {
	Name        string // 友链名称
	URL         string // 友链地址
	AvatarURL   string // 头像URL
	Description string // 友链简介
}

// SendFriendLinkNotificationByArgs 发送友链申请通知邮件。
// 参数说明：
//   - ctx: 上下文对象，用于控制请求的生命周期。
//   - email: 收件人的电子邮件地址。
//   - friendLink: 友链信息。
//   - smtpAccount: SMTP服务器的账户名。
//   - smtpAddress: SMTP服务器的地址。
//   - smtpAuthCode: SMTP服务器的授权码。
//   - smtpPort: SMTP服务器的端口号。
//
// 返回值：
//   - error: 如果发送邮件过程中发生错误，则返回错误信息；否则返回nil。
func SendFriendLinkNotificationByArgs(ctx context.Context, email string, friendLink FriendLinkData, smtpAccount, smtpAddress, smtpAuthCode string, smtpPort uint16) error {
	// 解析HTML模板，准备渲染友链信息。
	tmpl, err := template.New("friendlink").Parse(FriendLinkNotificationTemplate)
	if err != nil {
		return err
	}

	// 创建一个字符串构建器，用于存储渲染后的HTML内容。
	var htmlContent strings.Builder

	// 执行模板渲染，将友链信息插入到HTML模板中。
	err = tmpl.Execute(&htmlContent, FriendLinkData{
		Name:        template.HTMLEscapeString(friendLink.Name),
		URL:         template.HTMLEscapeString(friendLink.URL),
		AvatarURL:   template.HTMLEscapeString(friendLink.AvatarURL),
		Description: template.HTMLEscapeString(friendLink.Description),
	})
	if err != nil {
		return err
	}

	// 调用sendContent函数发送友链申请通知邮件。
	return sendContent(email, htmlContent.String(), FriendLinkNotificationSubject, smtpAccount, smtpAddress, smtpAuthCode, smtpPort)
}

// SendFriendLinkNotificationBySys 发送友链申请通知邮件给系统配置的邮箱地址。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据；
//   - friendLink: 友链信息；
//
// 返回值:
//   - error: 如果发送过程中出现错误，则返回具体的错误信息；否则返回 nil。
func SendFriendLinkNotificationBySys(ctx context.Context, friendLink FriendLinkData) error {
	// 调用 SendFriendLinkNotificationByArgs 函数发送友链申请通知邮件，
	// 使用系统配置中的 SMTP 账号、地址、授权码和端口信息。
	if err := SendFriendLinkNotificationByArgs(
		ctx,
		config.User.UserEmail,
		friendLink,
		config.Server.SmtpAccount,
		config.Server.SmtpAddress,
		config.Server.SmtpAuthCode,
		config.Server.SmtpPort,
	); err != nil {
		return err
	}

	return nil
}
