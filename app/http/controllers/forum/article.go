package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/spf13/cast"
)

func ArticleDetail(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	if id == 0 {
		renderNotFound(c)
		return
	}

	entity := articles.Get(id)
	if entity.Id == 0 {
		renderNotFound(c)
		return
	}
	loginUser := component.GetLoginUser(c)
	if entity.ProcessStatus != 0 && (loginUser == nil || !loginUser.IsAdmin) {
		renderNotFound(c)
		return
	}

	ensureRenderedHTML(&entity)
	props := buildArticleDetailProps(c, &entity)
	payload := PagePayload{
		Component: "article.detail",
		Props:     props,
		Meta:      buildArticleMeta(c, props.Article),
		Layout:    buildLayout(c, activeKeyForArticle(props.Article)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}

	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "article.gohtml", payload)
	articles.IncrementView(entity)
}

func renderNotFound(c *gin.Context) {
	payload := PagePayload{
		Component: "error.notFound",
		Props: ErrorPageProps{
			Code:    "404",
			Title:   "页面不存在",
			Message: "这个主题不存在，或已经被删除。",
		},
		Meta: PageMeta{
			Title:  pageTitle("页面不存在"),
			Robots: "noindex",
		},
		Layout:  buildLayout(c, "topics"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}

	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusNotFound)
	renderPage(c, "error.gohtml", payload)
}

func activeKeyForArticle(article ArticlePayload) string {
	if len(article.Categories) > 0 {
		return "category_" + cast.ToString(article.Categories[0].ID)
	}
	return "topics"
}
