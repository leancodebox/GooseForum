package controllers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"github.com/leancodebox/GooseForum/resource"
)

var activationTemplate = sync.OnceValues(func() (*template.Template, error) {
	return template.ParseFS(resource.GetTemplateFS(), "templates/view/activate.gohtml")
})

// ActivateAccount 激活处理函数
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
	user.Activate()
	if err = userservice.SaveUser(&user); err != nil {
		renderActivationPage(c, false, "激活失败")
		return
	}

	renderActivationPage(c, true, "账号激活成功")
}

// ActivateAccountData 激活页面数据
type ActivateAccountData struct {
	Title       string
	Status      string
	Message     string
	Success     bool
	Description string
}

// renderActivationPage 渲染账号激活结果页。
func renderActivationPage(c *gin.Context, success bool, message string) {
	status := "失败"
	description := "激活失败，请检查您的激活链接是否正确或联系管理员。"
	if success {
		status = "成功"
		description = "您的账号已成功激活！现在您可以使用完整的论坛功能，包括发帖、回复、个人中心等服务。"
	}

	tmpl, err := activationTemplate()
	if err != nil {
		slog.Error("parse activation template failed", "err", err)
		c.String(http.StatusInternalServerError, "activation page unavailable")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err = tmpl.Execute(c.Writer, struct {
		Data ActivateAccountData
		T    func(string, ...any) string
		Lang string
	}{
		Data: ActivateAccountData{
			Title:       fmt.Sprintf("账号激活%v", status),
			Status:      status,
			Message:     message,
			Success:     success,
			Description: description,
		},
		T:    activationText,
		Lang: "zh",
	}); err != nil {
		slog.Error("render activation template failed", "err", err)
	}
}

func activationText(key string, _ ...any) string {
	switch key {
	case "account_activation":
		return "账号激活"
	case "back_home":
		return "回到首页"
	case "contact_support":
		return "联系支持"
	case "login_now":
		return "立即登录"
	default:
		return key
	}
}
