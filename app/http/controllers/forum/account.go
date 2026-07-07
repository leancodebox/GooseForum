package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cast"
)

func Notifications(c *gin.Context) {
	payload := PagePayload{
		Component: "notifications.index",
		Props:     buildNotificationsPageProps(c),
		Meta:      buildSimpleMeta(c, "meta.notifications"),
		Layout:    buildLayout(c, "notifications"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "notifications.gohtml", payload)
}

func Messages(c *gin.Context) {
	payload := PagePayload{
		Component: "messages.index",
		Props:     buildMessagesPageProps(c),
		Meta:      buildSimpleMeta(c, "meta.messages"),
		Layout:    buildLayout(c, "messages"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "messages.gohtml", payload)
}

func Drafts(c *gin.Context) {
	payload := PagePayload{
		Component: "drafts.index",
		Props:     buildDraftsPageProps(c),
		Meta:      buildSimpleMeta(c, "meta.drafts"),
		Layout:    buildLayout(c, "drafts"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "drafts.gohtml", payload)
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
		Meta:      buildSimpleMeta(c, "meta.settings"),
		Layout:    buildLayout(c, "settings"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "settings.gohtml", payload)
}

func Publish(c *gin.Context) {
	topicID := cast.ToUint64(c.Query("id"))
	props, err := buildPublishPageProps(c, topicID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	title := "meta.publish"
	if topicID > 0 {
		title = "meta.editTopic"
	}
	payload := PagePayload{
		Component: "publish.index",
		Props:     props,
		Meta:      buildSimpleMeta(c, title),
		Layout:    buildLayout(c, "topics"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "publish.gohtml", payload)
}

func Login(c *gin.Context) {
	if component.LoginUserId(c) > 0 {
		c.Redirect(http.StatusFound, "/")
		return
	}

	payload := PagePayload{
		Component: "auth.login",
		Props:     buildLoginPageProps(c),
		Meta:      buildSimpleMeta(c, "meta.loginRegister"),
		Layout:    buildLayout(c, ""),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "login.gohtml", payload)
}

func ResetPassword(c *gin.Context) {
	payload := PagePayload{
		Component: "auth.resetPassword",
		Props:     ResetPasswordPageProps{Token: c.Query("token")},
		Meta:      buildSimpleMeta(c, "meta.resetPassword"),
		Layout:    buildLayout(c, ""),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "reset_password.gohtml", payload)
}

// buildSimpleMeta builds page metadata whose title is a translation key
// resolved in the request locale.
func buildSimpleMeta(c *gin.Context, titleKey string) PageMeta {
	return PageMeta{
		Title:     pageTitle(i18n.T(requestLang(c), titleKey)),
		Canonical: buildPageURL(c),
	}
}
