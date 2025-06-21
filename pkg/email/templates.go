package email

// VerificationCodeTemplate 验证码邮件的HTML模板
const VerificationCodeTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sparrow Blog 验证码</title>
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
            <h2 style="color: #3182ce;">Sparrow Blog</h2>
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

const (
	VerificationCodeSubject = "博客验证码"
)
