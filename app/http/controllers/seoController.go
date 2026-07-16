package controllers

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
)

//go:embed templ/robots.txt
var robotsTxt string

//go:embed templ/sitemap.xml.tmpl
var sitemapTpl string

const seoXMLCacheTTL = 10 * time.Second

var seoXMLCache = localcache.Cache[string]{MaxEntries: cacheconfig.Current().SEOXML}

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
	if !cacheableSEOHost(host) {
		xml, err := buildSitemapXML(host)
		if err != nil {
			c.String(http.StatusInternalServerError, "Sitemap build error")
			return
		}
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.Header("Cache-Control", "public, max-age=10")
		c.String(http.StatusOK, xml)
		return
	}
	xml, err := seoXMLCache.GetOrLoadE("sitemap:"+host, func() (string, error) {
		return buildSitemapXML(host)
	}, seoXMLCacheTTL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Sitemap build error")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=10")
	c.String(http.StatusOK, xml)
}

func buildSitemapXML(host string) (string, error) {
	list, _ := topics.GetLatestPublished(5000)

	tpl, err := template.New("sitemap").Parse(sitemapTpl)
	if err != nil {
		return "", err
	}

	var sitemaps []SitemapURL
	for _, topic := range list {
		sitemaps = append(sitemaps, SitemapURL{
			Loc:      host + urlconfig.PostDetail(topic.Id),
			Lastmod:  topic.UpdatedAt.Format(time.RFC3339),
			Priority: 0.7,
		})
	}
	categories := hotdataserve.GetCategory()
	for _, cat := range categories {
		sitemaps = append(sitemaps, SitemapURL{
			Loc:      host + fmt.Sprintf("/c/%s/%d", url.PathEscape(cat.Name), cat.Id),
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
	if !cacheableSEOHost(host) {
		rssString, err := buildRSSXML(host)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error generating RSS feed")
			return
		}
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.Header("Cache-Control", "public, max-age=10")
		c.String(http.StatusOK, rssString)
		return
	}
	rssString, err := seoXMLCache.GetOrLoadE("rss:"+host, func() (string, error) {
		return buildRSSXML(host)
	}, seoXMLCacheTTL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=10")
	c.String(http.StatusOK, rssString)
}

func cacheableSEOHost(host string) bool {
	if len(host) == 0 || len(host) > 128 {
		return false
	}
	if strings.ContainsAny(host, " \t\r\n") {
		return false
	}
	parsed, err := url.Parse(host)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func buildRSSXML(host string) (string, error) {
	settingConfig := hotdataserve.GetSiteSettingsConfigCache()
	topicList, err := topics.GetLatestPublished(100)
	if err != nil {
		return "", err
	}
	firstPostIDs := make([]uint64, 0, len(topicList))
	for _, topic := range topicList {
		if topic != nil && topic.FirstPostId > 0 {
			firstPostIDs = append(firstPostIDs, topic.FirstPostId)
		}
	}
	firstPostMap := posts.GetMapByIds(firstPostIDs)

	feedUpdated := time.Now()
	if len(topicList) > 0 {
		feedUpdated = topicList[0].UpdatedAt
		if feedUpdated.IsZero() {
			feedUpdated = topicList[0].CreatedAt
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

	for _, item := range topicList {
		firstPost := firstPostMap[item.FirstPostId]
		itemURL := host + urlconfig.PostDetail(item.Id)
		content := ""
		if firstPost != nil {
			content = firstPost.RenderedHTML
			if content == "" {
				content = markdown2html.MarkdownToHTML(firstPost.Content)
			}
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: itemURL},
			Description: item.Excerpt,
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
