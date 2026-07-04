package forum

import (
	"math"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
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
	lang := requestLang(c)
	title := i18n.T(lang, "search")
	if query != "" {
		title = query + " - " + title
	}
	return PageMeta{
		Title:       pageTitle(title),
		Description: i18n.T(lang, "meta.searchDesc", "site", siteTitle()),
		Canonical:   component.GetBaseUri(c) + buildSearchURL(query, 1),
	}
}

func totalPages(total int64, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}
