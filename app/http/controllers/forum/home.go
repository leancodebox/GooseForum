package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func Home(c *gin.Context) {
	sort, _ := lo.Coalesce(c.Query("sort"), "latest")
	page := lo.Ternary(cast.ToInt(c.Query("page")) <= 0, 1, cast.ToInt(c.Query("page")))

	topics := hotdataserve.GetLatestArticlesSimpleVoPaginated(page, sort)
	payload := PagePayload{
		Component: "home.index",
		Props:     buildHomeProps(page, sort, topics),
		Meta:      buildHomeMeta(c),
		Layout:    buildLayout(c, activeKeyForHome(sort)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}

	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "home.gohtml", payload)
}
