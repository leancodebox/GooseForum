package controllers

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/datacache"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
)

//go:embed templ/robots.txt
var robotsTxt string

//go:embed templ/sitemap.xml.tmpl
var sitemapTpl string

const seoXMLCacheTTL = 5 * time.Minute

var seoXMLCache = datacache.Cache[string]{}

// RenderRobotsTxt renders robots.txt.
func RenderRobotsTxt(c *gin.Context) {
	host := component.GetHost(c)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, fmt.Sprintf(robotsTxt, host, host))
}

type SitemapURL struct {
	Loc      string
	Lastmod  string
	Priority float64
}

// RenderSitemapXml renders sitemap.xml.
func RenderSitemapXml(c *gin.Context) {
	host := component.GetHost(c)
	xml, err := seoXMLCache.GetOrLoadE("sitemap:"+host, func() (string, error) {
		return buildSitemapXML(host)
	}, seoXMLCacheTTL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Sitemap build error")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=300")
	c.String(http.StatusOK, xml)
}

func buildSitemapXML(host string) (string, error) {
	list, _ := articles.GetLatestArticles(5000)

	tpl, err := template.New("sitemap").Parse(sitemapTpl)
	if err != nil {
		return "", err
	}

	var sitemaps []SitemapURL
	for _, article := range list {
		sitemaps = append(sitemaps, SitemapURL{
			Loc:      host + urlconfig.PostDetail(article.Id),
			Lastmod:  article.UpdatedAt.Format(time.RFC3339),
			Priority: 0.7,
		})
	}
	categories := hotdataserve.GetArticleCategory()
	for _, cat := range categories {
		sitemaps = append(sitemaps, SitemapURL{
			Loc:      host + fmt.Sprintf("/c/%s/%d", url.PathEscape(cat.Category), cat.Id),
			Lastmod:  cat.UpdatedAt.Format(time.RFC3339),
			Priority: 0.8,
		})
	}

	sitemaps = append(sitemaps, []SitemapURL{
		{
			Loc:      host + urlconfig.Links(),
			Priority: 0.8,
		},
		{
			Loc:      host + urlconfig.Home(),
			Priority: 1,
		}}...)

	data := gin.H{
		"Sitemaps": sitemaps,
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderRss(c *gin.Context) {
	host := component.GetHost(c)
	rssString, err := seoXMLCache.GetOrLoadE("rss:"+host, func() (string, error) {
		return buildRSSXML(host)
	}, seoXMLCacheTTL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=300")
	c.String(http.StatusOK, rssString)
}

func buildRSSXML(host string) (string, error) {
	settingConfig := hotdataserve.GetSiteSettingsConfigCache()
	articleList, err := articles.GetLatestArticlesWithContent(100)
	if err != nil {
		return "", err
	}

	feedUpdated := time.Now()
	if len(articleList) > 0 {
		feedUpdated = articleList[0].UpdatedAt
		if feedUpdated.IsZero() {
			feedUpdated = articleList[0].CreatedAt
		}
	}

	feed := &feeds.Feed{
		Title:       settingConfig.SiteName,
		Link:        &feeds.Link{Href: host},
		Description: settingConfig.SiteDescription,
		Author:      &feeds.Author{Name: settingConfig.SiteName, Email: settingConfig.SiteEmail},
		Created:     feedUpdated,
		Updated:     feedUpdated,
	}

	for _, item := range articleList {
		itemURL := host + urlconfig.PostDetail(item.Id)
		content := item.RenderedHTML
		if content == "" {
			content = markdown2html.MarkdownToHTML(item.Content)
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: itemURL},
			Description: item.Description,
			Content:     content,
			Id:          itemURL,
			Created:     item.CreatedAt,
			Updated:     item.UpdatedAt,
		})
	}

	rssString, err := feed.ToRss()
	if err != nil {
		return "", err
	}

	return rssString, nil
}
