package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ManageHomeProps struct {
	LegacyAdminURL string `json:"legacyAdminUrl"`
}

func Manage(c *gin.Context) {
	payload := PagePayload{
		Component: "admin.shell",
		Props: ManageHomeProps{
			LegacyAdminURL: "/admin/",
		},
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
	renderPage(c, "manage.gohtml", payload)
}
