package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func Category(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	category := hotdataserve.GetCategoryById(id)
	if category == nil {
		renderNotFoundWithMessage(c, "这个分类不存在，或已经被删除。")
		return
	}

	sort, _ := lo.Coalesce(c.Param("sort"), "latest")
	page := lo.Ternary(cast.ToInt(c.Query("page")) <= 0, 1, cast.ToInt(c.Query("page")))
	topics := hotdataserve.GetArticlesByCategorySimpleVo(id, sort, page)

	payload := PagePayload{
		Component: "category.index",
		Props:     buildCategoryPageProps(category, page, sort, topics),
		Meta:      buildCategoryMeta(c, category),
		Layout:    buildLayout(c, "category_"+cast.ToString(category.Id)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "category.gohtml", payload)
}
