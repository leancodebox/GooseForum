package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/service/themeservice"
)

type ThemePreviewProps struct {
	Theme    pageConfig.SiteThemeConfig `json:"theme"`
	Defaults pageConfig.SiteThemeConfig `json:"defaults"`
}

func ThemePreview(c *gin.Context) {
	payload := PagePayload{
		Component: "theme.preview",
		Props: ThemePreviewProps{
			Theme:    themeservice.LoadConfig(),
			Defaults: themeservice.Defaults(),
		},
		Meta:    buildSimpleMeta(c, "主题预览"),
		Layout:  buildLayout(c, "theme-preview"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "theme.gohtml", payload)
}

func SiteThemeCSS(c *gin.Context) {
	runtimeTheme := themeservice.Runtime()
	if !runtimeTheme.Enabled || runtimeTheme.CSS == "" {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Content-Type", "text/css; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("ETag", runtimeTheme.ETag)
	if match := c.GetHeader("If-None-Match"); match != "" && match == runtimeTheme.ETag {
		c.Status(http.StatusNotModified)
		return
	}
	c.String(http.StatusOK, runtimeTheme.CSS)
}
