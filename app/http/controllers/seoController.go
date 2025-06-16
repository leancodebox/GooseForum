package controllers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/spf13/cast"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
)

// RenderRobotsTxt 渲染 robots.txt
func RenderRobotsTxt(c *gin.Context) {
	host := getHost(c)
	robotsTxt := fmt.Sprintf(`User-agent: *
Allow: /
Allow: /post
Allow: /post/*

# Sitemaps
Sitemap: %s/sitemap.xml

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
		<lastmod>%v</lastmod>
        <priority>0.7</priority>
    </url>
`, host, item.Id, item.UpdatedAt.Format(time.RFC3339)))
	}

	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1" xmlns:video="http://www.google.com/schemas/sitemap-video/1.1" xmlns:news="http://www.google.com/schemas/sitemap-news/0.9">
	%v
    <url>
        <loc>%s/post</loc>
        <priority>0.8</priority>
    </url>
	<url>
        <loc>%s/links</loc>
        <priority>1.0</priority>
    </url>
    <url>
        <loc>%s/</loc>
        <priority>1.0</priority>
    </url>
</urlset>`, sb.String(), host, host, host)

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, sitemap)
}

var formatRFC822WithZone = "02 Jan 06 15:04 -0700"

func RenderRssV2(c *gin.Context) {
	host := getHost(c)
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
		"GooseForum - 最新文章",
		host+"/rss.xml",
		"GooseForum",
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
