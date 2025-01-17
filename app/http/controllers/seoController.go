package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RenderRobotsTxt 渲染 robots.txt
func RenderRobotsTxt(c *gin.Context) {
	host := c.Request.Host
	robotsTxt := fmt.Sprintf(`User-agent: *
Allow: /
Allow: /articles
Allow: /articles/*

# Sitemaps
Sitemap: https://%s/sitemap.xml

# Crawl-delay
Crawl-delay: 10

# Disallow
Disallow: /api/
Disallow: /admin/
Disallow: /actor/
`, host)

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, robotsTxt)
}

// RenderSitemapXml 渲染 sitemap.xml
func RenderSitemapXml(c *gin.Context) {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>%s/</loc>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
    <url>
        <loc>%s/articles</loc>
        <changefreq>hourly</changefreq>
        <priority>0.9</priority>
    </url>
</urlset>`, host, host)

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, sitemap)
}
