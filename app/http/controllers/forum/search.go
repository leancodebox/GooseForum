package forum

import (
	"math"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

func Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	props := buildSearchPageProps(query, page)
	payload := PagePayload{
		Component: "search.index",
		Props:     props,
		Meta:      buildSearchMeta(c, query),
		Layout:    buildLayout(c, "search"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "search.gohtml", payload)
}

func buildSearchMeta(c *gin.Context, query string) PageMeta {
	title := "搜索"
	if query != "" {
		title = query + " - 搜索"
	}
	return PageMeta{
		Title:       pageTitle(title),
		Description: "搜索 " + siteTitle() + " 主题、关键词和讨论。",
		Canonical:   component.GetBaseUri(c) + buildSearchURL(query, 1),
	}
}

func totalPages(total int64, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}
