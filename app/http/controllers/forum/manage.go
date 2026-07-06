package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
)

type ManageHomeProps struct{}

func Manage(c *gin.Context) {
	payload := PagePayload{
		Component: "admin.shell",
		Props:     ManageHomeProps{},
		Meta: PageMeta{
			Title:  pageTitle(i18n.T(requestLang(c), "meta.admin")),
			Robots: "noindex,nofollow",
		},
		Layout:  buildLayout(c, "manage"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}

	renderPage(c, "admin.gohtml", payload)
}
