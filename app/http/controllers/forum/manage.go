package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ManageHomeProps struct{}

func Manage(c *gin.Context) {
	payload := PagePayload{
		Component: "admin.shell",
		Props:     ManageHomeProps{},
		Meta: PageMeta{
			Title:  pageTitle("管理后台"),
			Robots: "noindex,nofollow",
		},
		Layout:  buildLayout(c, "manage"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}

	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "admin.gohtml", payload)
}
