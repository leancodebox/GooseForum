package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
)

// 添加激活处理函数
func ActivateAccount(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		renderActivationPage(c, false, "无效的激活链接")
		return
	}

	// 解析激活令牌
	claims, err := tokenservice.ParseActivationToken(token)
	if err != nil {
		renderActivationPage(c, false, "激活链接已过期或无效")
		return
	}

	// 获取用户信息
	user, err := users.Get(claims.UserId)
	if err != nil {
		renderActivationPage(c, false, "用户不存在")
		return
	}

	// 检查邮箱是否匹配
	if user.Email != claims.Email {
		renderActivationPage(c, false, "激活链接无效")
		return
	}

	// 激活账号
	err = user.Activate()
	if err != nil {
		renderActivationPage(c, false, "激活失败")
		return
	}

	renderActivationPage(c, true, "账号激活成功")
}

// 添加新的辅助函数
func renderActivationPage(c *gin.Context, success bool, message string) {
	status := "失败"
	if success {
		status = "成功"
	}

	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>账号激活` + status + ` - Goose Forum</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f0f2f5;
            color: #333;
        }
        .container {
            text-align: center;
            padding: 2.5rem;
            background-color: white;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            max-width: 90%;
            width: 420px;
        }
        .icon {
            font-size: 48px;
            margin-bottom: 1rem;
        }
        .title {
            font-size: 24px;
            font-weight: 600;
            margin: 1rem 0;
            color: #1a1a1a;
        }
        .message {
            margin: 1rem 0;
            line-height: 1.6;
            color: ` + func() string {
		if success {
			return "#4caf50"
		}
		return "#f44336"
	}() + `;
            font-size: 16px;
        }
        .description {
            color: #666;
            margin: 1rem 0;
            font-size: 14px;
            line-height: 1.6;
        }
        .button-group {
            margin-top: 1.5rem;
            display: flex;
            gap: 1rem;
            justify-content: center;
        }
        .home-link {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            background-color: #1890ff;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            transition: all 0.3s ease;
        }
        .home-link:hover {
            background-color: #096dd9;
            transform: translateY(-1px);
        }
        .login-link {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            background-color: #f0f0f0;
            color: #333;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            transition: all 0.3s ease;
        }
        .login-link:hover {
            background-color: #d9d9d9;
            transform: translateY(-1px);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">` + func() string {
		if success {
			return "✅"
		}
		return "❌"
	}() + `</div>
        <h2 class="title">账号激活` + status + `</h2>
        <p class="message">` + message + `</p>
        <p class="description">` + func() string {
		if success {
			return "您的账号已成功激活！现在您可以使用完整的论坛功能，包括发帖、回复、个人中心等服务。"
		}
		return "激活失败可能是因为链接已过期或无效。您可以尝试重新注册或联系管理员获取帮助。"
	}() + `</p>
        <div class="button-group">
            <a href="/" class="home-link">返回首页</a>
            ` + func() string {
		if success {
			return `<a href="/login" class="login-link">立即登录</a>`
		}
		return `<a href="/login?model=register" class="login-link">重新注册</a>`
	}() + `
        </div>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, html)
}
