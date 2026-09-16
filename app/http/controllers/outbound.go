package controllers

import (
	_ "embed"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/outbound"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

//go:embed templ/outbound.gohtml
var outboundHTML string
var outboundTemplate = template.Must(template.New("outbound").Parse(outboundHTML))

func Outbound(c *gin.Context) {
	site := hotdataserve.GetSiteSettingsConfigCache()
	config := hotdataserve.GetPostingSettingsConfigCache().ExternalLinks
	serveOutbound(c, config.Enabled, config.Whitelist, site.SiteUrl, site.SiteName, hotdataserve.GetSiteChromeConfigCache())
}

func serveOutbound(c *gin.Context, enabled bool, whitelist []string, siteURL, siteName string, chrome pageConfig.SiteChromeConfig) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Robots-Tag", "noindex, nofollow")
	c.Header("Content-Security-Policy", "default-src 'none'; img-src 'self' https: http:; style-src 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")
	destination, err := outbound.Parse(c.Query("url"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid external URL")
		return
	}
	base, _ := url.Parse(siteURL)
	internal := base != nil && base.Host != "" && strings.EqualFold(base.Host, destination.Host)
	if !enabled || internal || outbound.Allowed(destination, whitelist) {
		c.Redirect(http.StatusFound, destination.String())
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_ = outboundTemplate.Execute(c.Writer, struct {
		Site, URL, Host                  string
		BrandType, BrandText, BrandImage string
	}{siteName, destination.String(), destination.Host, chrome.BrandType, chrome.BrandText, chrome.BrandImage})
}
