package controllers

import (
	"html/template"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
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
		renderActivationPage(c, false, "activationInvalidLink")
		return
	}

	// 解析激活令牌
	claims, err := tokenservice.ParseActivationToken(token)
	if err != nil {
		renderActivationPage(c, false, "activationExpired")
		return
	}

	// 获取用户信息
	user, err := users.Get(claims.UserId)
	if err != nil {
		renderActivationPage(c, false, "activationUserNotFound")
		return
	}

	// 检查邮箱是否匹配
	if user.Email != claims.Email {
		renderActivationPage(c, false, "activationLinkInvalid")
		return
	}

	// 激活账号
	user.Activate()
	if err = userservice.SaveUser(&user); err != nil {
		renderActivationPage(c, false, "activationFailed")
		return
	}

	renderActivationPage(c, true, "activationSuccess")
}

// ActivateAccountData 激活页面数据
type ActivateAccountData struct {
	Title       string
	Message     string
	Success     bool
	Description string
}

// renderActivationPage 渲染账号激活结果页。messageKey 为 i18n 文案键。
func renderActivationPage(c *gin.Context, success bool, messageKey string) {
	lang := component.RequestLang(c)
	tr := i18n.Func(lang)

	title := tr("activationTitleFail")
	description := tr("activationDescFail")
	if success {
		title = tr("activationTitleSuccess")
		description = tr("activationDescSuccess")
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
			Title:       title,
			Message:     tr(messageKey),
			Success:     success,
			Description: description,
		},
		T:    tr,
		Lang: lang,
	}); err != nil {
		slog.Error("render activation template failed", "err", err)
	}
}
