package forum

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

// RenderErrorPage renders the shared site error page for browser-facing flows
// that live outside the forum package, such as OAuth callbacks.
func RenderErrorPage(c *gin.Context, status int, title string, messageCode component.MessageCode, params component.MessageParams) {
	payload := PagePayload{
		Component: "error.generic",
		Props: ErrorPageProps{
			Code:        strconv.Itoa(status),
			Title:       title,
			MessageCode: messageCode,
			Params:      params,
		},
		Meta: PageMeta{
			Title:  pageTitle(title),
			Robots: "noindex",
		},
		Layout:  buildLayout(c, ""),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}

	renderPageWithStatus(c, status, "error.gohtml", payload)
}

func RenderOAuthErrorPage(c *gin.Context, status int, messageCode component.MessageCode) {
	RenderErrorPage(c, status, "OAuth", messageCode, nil)
}

func RenderInternalOAuthErrorPage(c *gin.Context, messageCode component.MessageCode) {
	RenderOAuthErrorPage(c, http.StatusInternalServerError, messageCode)
}

func RenderNotFoundPage(c *gin.Context, messageCode component.MessageCode) {
	payload := PagePayload{
		Component: "error.notFound",
		Props: ErrorPageProps{
			Code:        "404",
			Title:       "页面不存在",
			MessageCode: messageCode,
		},
		Meta: PageMeta{
			Title:  pageTitle("页面不存在"),
			Robots: "noindex",
		},
		Layout:  buildLayout(c, ""),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}

	renderPageWithStatus(c, http.StatusNotFound, "error.gohtml", payload)
}
