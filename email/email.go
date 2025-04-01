package email

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/utils"
	"h2blog_server/storage"
	"html/template"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

// SendVerificationCodeEmail 发送验证邮件到指定邮箱。
// 该函数首先生成一个验证代码，然后根据当前环境（配置服务器环境或运行时环境）存储或获取验证代码，
// 最后构建并发送包含验证代码的邮件到指定邮箱。
// 参数:
//
//	ctx - 上下文，用于传递请求范围的数据。
//	email - 接收验证邮件的邮箱地址。
//
// 返回值:
//
//	如果发送邮件过程中发生错误，则返回该错误；否则返回nil。
func SendVerificationCodeEmail(ctx context.Context, email string) error {
	// 生成验证代码
	code, err := utils.HashWithLength(config.User.UserEmail+time.Now().String(), 20)
	if err != nil {
		return err
	}

	// 根据当前环境处理验证代码
	switch env.CurrentEnv {
	case env.InitializedEnv:
		// 在配置服务器环境中，如果环境变量中没有验证代码，则设置为当前生成的代码
		if env.VerificationCode == "" {
			env.VerificationCode = code
		} else {
			// 如果环境变量中已有验证代码，则使用它
			code = env.VerificationCode
		}
	case env.ProvEnv, env.DebugEnv:
		// 在运行时环境中，尝试从缓存中获取验证代码
		c, err := storage.Storage.Cache.GetString(ctx, storage.VerificationCodeKey)
		if err != nil {
			// 如果缓存中没有验证代码，将其存储到缓存中，设置过期时间为5分钟
			err = storage.Storage.Cache.SetWithExpired(ctx, storage.VerificationCodeKey, code, 5*time.Minute)
			if err != nil {
				msg := fmt.Sprintf("缓存验证码失败: %v", err)
				return errors.New(msg)
			}
		} else {
			code = c
		}
	}

	// 创建邮件内容
	m := gomail.NewMessage()
	// 发件人
	m.SetHeader("From", config.User.SmtpAccount)
	// 收件人
	m.SetHeader("To", email)
	// 主题
	m.SetHeader("Subject", "博客验证码")

	// HTML 正文模板
	htmlTemplate := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>H2Blog 验证码</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: #f5f7fa;
            color: #333;
            margin: 0;
            padding: 0;
            -webkit-font-smoothing: antialiased;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background: linear-gradient(135deg, #ffffff, #f5f7fa);
            padding: 40px 30px;
            border-radius: 16px;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.05);
            border: 1px solid rgba(0, 0, 0, 0.05);
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo img {
            height: 50px;
        }
        h1 {
            color: #2d3748;
            text-align: center;
            font-size: 24px;
            font-weight: 600;
            margin-bottom: 30px;
        }
        .code-container {
            background-color: #f8fafc;
            border: 1px dashed #cbd5e0;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            margin: 25px 0;
        }
        .verification-code {
            font-family: 'Courier New', monospace;
            font-size: 28px;
            font-weight: bold;
            color: #3182ce;
            letter-spacing: 2px;
            word-break: break-all;
            line-height: 1.4;
            text-align: center;
        }
        .message {
            text-align: center;
            color: #718096;
            font-size: 16px;
            margin: 25px 0;
            line-height: 1.6;
        }
        @media (max-width: 600px) {
            .container {
                margin: 20px auto;
                padding: 25px 15px;
            }
            .verification-code {
                font-size: 22px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <!-- 替换为实际的博客logo -->
            <h2 style="color: #3182ce;">H2Blog</h2>
        </div>
        
        <h1>验证您的邮箱</h1>
        
        <div class="message">
            请使用以下验证码完成邮箱验证流程：
        </div>
        
        <div class="code-container">
            <div class="verification-code">{{.Code}}</div>
        </div>
        
        <div class="message">
            <strong>此验证码将在 5 分钟内有效</strong><br>
            请勿将验证码泄露给他人，以确保您的账户安全。
        </div>
    </div>
</body>
</html>
	`

	// 解析模板
	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	// 创建一个字符串构建器来存储渲染后的HTML内容
	var htmlContent strings.Builder

	// 执行模板并写入构建器
	err = tmpl.Execute(&htmlContent, struct {
		Code string
	}{Code: template.HTMLEscapeString(code)})
	if err != nil {
		return err
	}

	// 设置邮件正文
	m.SetBody("text/html", htmlContent.String())

	// 配置 SMTP 服务器
	d := gomail.NewDialer(config.User.SmtpAddress, int(config.User.SmtpPort), config.User.SmtpAccount, config.User.SmtpAuthCode)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
