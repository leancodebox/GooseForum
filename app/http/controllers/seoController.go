package controllers

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/spf13/cast"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
)

//go:embed templ/robots.txt
var robotsTxt string

//go:embed templ/sitemap.xml.tmpl
var sitemapTpl string

// RenderRobotsTxt 渲染 robots.txt
func RenderRobotsTxt(c *gin.Context) {
	host := component.GetHost(c)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, fmt.Sprintf(robotsTxt, host))
}

type SitemapURL struct {
	Loc      string
	Lastmod  string
	Priority float64
}

// RenderSitemapXml 渲染 sitemap.xml
func RenderSitemapXml(c *gin.Context) {
	host := component.GetHost(c)
	list, _ := articles.GetLatestArticles(160)

	tpl, err := template.New("sitemap").Parse(sitemapTpl)
	if err != nil {
		c.String(http.StatusInternalServerError, "Sitemap template parse error")
		return
	}

	var sitemaps []SitemapURL
	for _, article := range list {
		sitemaps = append(sitemaps, SitemapURL{
			Loc:      fmt.Sprintf("%s/post/%d", host, article.Id),
			Lastmod:  article.UpdatedAt.Format(time.RFC3339),
			Priority: 0.7,
		})
	}
	sitemaps = append(sitemaps, []SitemapURL{
		{
			Loc:      host + "/post",
			Priority: 0.8,
		},
		{
			Loc:      host + "/links",
			Priority: 0.8,
		},
		{
			Loc:      host + "/",
			Priority: 1,
		}}...)

	data := gin.H{
		"Sitemaps": sitemaps,
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		c.String(http.StatusInternalServerError, "Sitemap template execute error")
		return
	}

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, buf.String())
}

func RenderRssV2(c *gin.Context) {
	settingConfig := hotdataserve.GetSiteSettingsConfigCache()
	host := component.GetHost(c)
	articleList, err := articles.GetLatestArticles(100)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}

	// 创建Feed对象
	feed := &feeds.Feed{
		Title:       settingConfig.SiteName,
		Link:        &feeds.Link{Href: host},
		Description: settingConfig.SiteDescription,
		Author:      &feeds.Author{Name: settingConfig.SiteName, Email: settingConfig.SiteEmail},
		Created:     time.Now(),
	}

	// 添加文章项
	for _, item := range articleList {
		// 使用RenderedHTML作为内容，如果为空则使用Description

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/post/%d", host, item.Id)},
			Description: item.Description,
			Id:          cast.ToString(item.Id),
			Created:     item.CreatedAt,
		})
	}

	// 生成RSS XML
	rssString, err := feed.ToRss()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, rssString)
}
