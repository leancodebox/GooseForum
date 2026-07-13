package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func Category(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	category := hotdataserve.GetCleanCategoryById(id)
	if category == nil {
		renderNotFoundWithMessage(c, component.MessagePageNotFound)
		return
	}

	sort, _ := lo.Coalesce(c.Param("sort"), "latest")
	page := lo.Ternary(cast.ToInt(c.Query("page")) <= 0, 1, cast.ToInt(c.Query("page")))
	topicPage := hotdataserve.GetTopicsByCategorySimpleVo(id, sort, page)

	payload := PagePayload{
		Component: "category.index",
		Props:     buildCategoryPageProps(category, page, sort, topicPage.Topics, topicPage.HasNext),
		Meta:      buildCategoryMeta(c, category, page, sort, topicPage.HasNext),
		Layout:    buildLayout(c, "category_"+cast.ToString(category.Id)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "category.gohtml", payload)
}
