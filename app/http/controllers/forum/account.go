package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cast"
)

func Notifications(c *gin.Context) {
	payload := PagePayload{
		Component: "notifications.index",
		Props:     buildNotificationsPageProps(c),
		Meta:      buildSimpleMeta(c, "通知"),
		Layout:    buildLayout(c, "notifications"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "notifications.gohtml", payload)
}

func Messages(c *gin.Context) {
	payload := PagePayload{
		Component: "messages.index",
		Props:     buildMessagesPageProps(c),
		Meta:      buildSimpleMeta(c, "私信"),
		Layout:    buildLayout(c, "messages"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "messages.gohtml", payload)
}

func Settings(c *gin.Context) {
	user, err := users.Get(component.LoginUserId(c))
	if err != nil || user.Id == 0 {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}
	payload := PagePayload{
		Component: "settings.index",
		Props:     buildSettingsPageProps(user),
		Meta:      buildSimpleMeta(c, "个人设置"),
		Layout:    buildLayout(c, "settings"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "settings.gohtml", payload)
}

func Publish(c *gin.Context) {
	articleID := cast.ToUint64(c.Query("id"))
	props, err := buildPublishPageProps(c, articleID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	title := "发布主题"
	if articleID > 0 {
		title = "编辑主题"
	}
	payload := PagePayload{
		Component: "publish.index",
		Props:     props,
		Meta:      buildSimpleMeta(c, title),
		Layout:    buildLayout(c, "topics"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "publish.gohtml", payload)
}

func Login(c *gin.Context) {
	payload := PagePayload{
		Component: "auth.login",
		Props:     buildLoginPageProps(c),
		Meta:      buildSimpleMeta(c, "登录/注册"),
		Layout:    buildLayout(c, ""),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "login.gohtml", payload)
}

func ResetPassword(c *gin.Context) {
	payload := PagePayload{
		Component: "auth.resetPassword",
		Props:     ResetPasswordPageProps{Token: c.Query("token")},
		Meta:      buildSimpleMeta(c, "重置密码"),
		Layout:    buildLayout(c, ""),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "reset_password.gohtml", payload)
}

func buildSimpleMeta(c *gin.Context, title string) PageMeta {
	return PageMeta{
		Title:     pageTitle(title),
		Canonical: buildPageURL(c),
	}
}
