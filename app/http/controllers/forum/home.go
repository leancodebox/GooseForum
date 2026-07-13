package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func Home(c *gin.Context) {
	sort, _ := lo.Coalesce(c.Query("sort"), "latest")
	page := lo.Ternary(cast.ToInt(c.Query("page")) <= 0, 1, cast.ToInt(c.Query("page")))

	topicPage := hotdataserve.GetLatestTopicsSimpleVoPaginated(page, sort)
	payload := PagePayload{
		Component: "home.index",
		Props:     buildHomeProps(page, sort, topicPage.Topics, topicPage.HasNext),
		Meta:      buildHomeMeta(c, page, sort, topicPage.HasNext),
		Layout:    buildLayout(c, activeKeyForHome(sort)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}

	renderPage(c, "home.gohtml", payload)
}
