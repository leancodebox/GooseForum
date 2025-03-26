package controllers

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
)

// RenderRobotsTxt 渲染 robots.txt
func RenderRobotsTxt(c *gin.Context) {
	host := c.Request.Host
	robotsTxt := fmt.Sprintf(`User-agent: *
Allow: /
Allow: /post
Allow: /post/*

# Sitemaps
Sitemap: https://%s/sitemap.xml

# Disallow
Disallow: /api/
Disallow: /admin/
Disallow: /actor/
Disallow: /app/
`, host)

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, robotsTxt)
}

// RenderSitemapXml 渲染 sitemap.xml
func RenderSitemapXml(c *gin.Context) {
	host := getHost(c)
	sb := strings.Builder{}
	list, _ := articles.GetLatestArticles(160)
	for _, item := range list {
		sb.WriteString(fmt.Sprintf(`    <url>
        <loc>%v/post/%v</loc>
        <changefreq>daily</changefreq>
        <priority>0.9</priority>
    </url>
`, host, item.Id))
	}

	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1" xmlns:video="http://www.google.com/schemas/sitemap-video/1.1" xmlns:news="http://www.google.com/schemas/sitemap-news/0.9">
	%v
    <url>
        <loc>%s/</loc>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
    <url>
        <loc>%s/post</loc>
        <changefreq>hourly</changefreq>
        <priority>0.9</priority>
    </url>
</urlset>`, sb.String(), host, host)

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, sitemap)
}

// RenderRssFeed 渲染 RSS feed
func RenderRssFeed(c *gin.Context) {
	host := getHost(c)

	// 获取最新的文章列表
	articleList, err := articles.GetLatestArticles(20)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}

	// 生成RSS内容
	now := time.Now().Format(time.RFC1123Z)
	rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
    <title><![CDATA[GooseForum - 最新文章]]></title>
    <link>%s</link>
    <description><![CDATA[GooseForum的最新文章和讨论]]></description>
    <language>zh-CN</language>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href="%s/rss.xml" rel="self" type="application/rss+xml" />
`, html.EscapeString(host), now, html.EscapeString(host))

	// 添加文章条目
	for _, article := range articleList {
		pubDate := article.CreatedAt.Format(time.RFC1123Z)
		rss += fmt.Sprintf(`
    <item>
        <title><![CDATA[%s]]></title>
        <link>%s</link>
        <guid>%s</guid>
        <pubDate>%s</pubDate>
        <description><![CDATA[%s]]></description>
        <author><![CDATA[%s]]></author>
        <category><![CDATA[%s]]></category>
    </item>`,
			article.Title,
			html.EscapeString(fmt.Sprintf("%s/articles/%d", host, article.Id)),
			html.EscapeString(fmt.Sprintf("%s/articles/%d", host, article.Id)),
			pubDate,
			"",
			"author",   // 这里可以添加实际作者信息
			"category", // 这里可以添加实际分类信息
		)
	}

	rss += `
</channel>
</rss>`

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, rss)
}
