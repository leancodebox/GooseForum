package controllers

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"fmt"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/spf13/cast"
	"html"
	"net/http"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
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

var formatRFC822WithZone = "02 Jan 06 15:04 -0700"

func RenderRssV2(c *gin.Context) {
	settingConfig := hotdataserve.GetSiteSettingsConfigCache()
	host := component.GetHost(c)
	articleList, err := articles.GetLatestArticles(100)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}

	articlesList := array.Map(articleList, func(item articles.SmallEntity) Item {
		return Item{
			Title:       item.Title,
			Link:        html.EscapeString(fmt.Sprintf("%s/post/%d", host, item.Id)),
			Description: CDATA(item.Description),
			PubDate:     item.CreatedAt.Format(formatRFC822WithZone),
			GUID:        cast.ToString(item.Id),
		}
	})

	rss, err := GenerateRSS(
		settingConfig.SiteName,
		host+"/rss.xml",
		settingConfig.SiteDescription,
		articlesList,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, rss)
}

// RSS 2.0规范核心结构体
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description CDATA  `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"` // 可选唯一标识符
}

type CDATA string

func (c CDATA) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		string `xml:",cdata"`
	}{string(c)}, start)
}

// GenerateRSS 生成符合RSS 2.0规范的XML字符串
func GenerateRSS(title, link, description string, items []Item) (string, error) {
	// 构建Channel结构
	channel := Channel{
		Title:       title,
		Link:        link,
		Description: description,
		PubDate:     time.Now().Format(formatRFC822WithZone),
	}

	// 转换Items
	for _, item := range items {
		channel.Items = append(channel.Items, item)
	}

	// 组装完整RSS结构
	rss := RSS{
		Version: "2.0",
		Channel: channel,
	}

	// 生成XML
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(rss); err != nil {
		return "", err
	}
	return buf.String(), nil
}
