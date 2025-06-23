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

// FriendLinkNotificationTemplate 友链申请通知邮件的HTML模板
const FriendLinkNotificationTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sparrow Blog 友链通知</title>
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
        .friend-link-container {
            background-color: #f8fafc;
            border: 1px solid #e2e8f0;
            border-radius: 12px;
            padding: 25px;
            margin: 25px 0;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
        }
        .friend-link-header {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
        }
        .friend-link-avatar {
            width: 60px;
            height: 60px;
            border-radius: 50%;
            object-fit: cover;
            border: 3px solid #3182ce;
            margin-right: 15px;
        }
        .friend-link-name {
            font-size: 20px;
            font-weight: 600;
            color: #2d3748;
            margin: 0;
        }
        .friend-link-url {
            font-size: 14px;
            color: #3182ce;
            text-decoration: none;
            word-break: break-all;
        }
        .friend-link-url:hover {
            text-decoration: underline;
        }
        .friend-link-description {
            background-color: #ffffff;
            padding: 15px;
            border-radius: 8px;
            color: #4a5568;
            line-height: 1.6;
            border-left: 4px solid #3182ce;
            margin-top: 15px;
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
            .friend-link-header {
                flex-direction: column;
                text-align: center;
            }
            .friend-link-avatar {
                margin-right: 0;
                margin-bottom: 10px;
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
        
        <h1>📝 收到新的友链申请</h1>
        
        <div class="message">
            您好！有一个新的友链申请需要您的审核：
        </div>
        
        <div class="friend-link-container">
            <div class="friend-link-header">
                <img src="{{.AvatarURL}}" alt="{{.Name}}" class="friend-link-avatar" onerror="this.src='data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPGNpcmNsZSBjeD0iMzAiIGN5PSIzMCIgcj0iMzAiIGZpbGw9IiMzMTgyY2UiLz4KPHN2ZyB4PSIxNSIgeT0iMTUiIHdpZHRoPSIzMCIgaGVpZ2h0PSIzMCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJ3aGl0ZSI+CjxwYXRoIGQ9Ik0xMiAyQzEzLjEgMiAxNCAyLjkgMTQgNEMxNCA1LjEgMTMuMSA2IDEyIDZDMTAuOSA2IDEwIDUuMSAxMCA0QzEwIDIuOSAxMC45IDIgMTIgMlpNMjEgOVYyMkgxNVYxNkgxM1YyMkg3VjlIMjFaTTkgN0M5IDcuNiA5LjQgOCAxMCA4SDE0QzE0LjYgOCAxNSA3LjYgMTUgN0MxNSA2LjQgMTQuNiA2IDE0IDZIMTBDOS40IDYgOSA2LjQgOSA3WiIvPgo8L3N2Zz4KPC9zdmc+'">
                <div>
                    <h3 class="friend-link-name">{{.Name}}</h3>
                    <a href="{{.URL}}" class="friend-link-url" target="_blank">{{.URL}}</a>
                </div>
            </div>
            
            {{if .Description}}
            <div class="friend-link-description">
                <strong>站点简介：</strong><br>
                {{.Description}}
            </div>
            {{end}}
        </div>
        
        <div class="message">
            请登录管理后台查看并处理该友链申请。<br>
            <strong>友链交换让我们的网络世界更加精彩！🌟</strong>
        </div>
    </div>
</body>
</html>
`

const (
	VerificationCodeSubject       = "博客验证码"
	FriendLinkNotificationSubject = "新的友链申请"
)
