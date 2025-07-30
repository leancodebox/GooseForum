package viewrender

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"html/template"
	"strings"
	"time"
)

type PageMetaBuilder struct {
	meta       *PageMeta
	siteConfig *pageConfig.SiteSettingsConfig
}

func NewPageMetaBuilder() *PageMetaBuilder {
	siteConfig := hotdataserve.GetSiteSettingsConfigCache()

	return &PageMetaBuilder{
		meta: &PageMeta{
			Favicon:    siteConfig.SiteLogo,
			ThemeColor: "#1976d2",
			OG: OpenGraphMeta{
				Type:     "website",
				SiteName: siteConfig.SiteName,
			},
			Twitter: TwitterCardMeta{
				Card: "summary_large_image",
				Site: "@GooseForum",
			},
		},
		siteConfig: &siteConfig,
	}
}

// 基础信息设置
func (b *PageMetaBuilder) SetTitle(title string) *PageMetaBuilder {
	b.meta.Title = title
	if b.meta.OG.Title == "" {
		b.meta.OG.Title = title
	}

	if b.meta.Twitter.Title == "" {
		b.meta.Twitter.Title = title
	}
	return b
}

func (b *PageMetaBuilder) SetDescription(desc string) *PageMetaBuilder {
	b.meta.Description = desc
	if b.meta.OG.Description == "" {
		b.meta.OG.Description = desc
	}
	if b.meta.Twitter.Description == "" {
		b.meta.Twitter.Description = desc
	}
	return b
}

func (b *PageMetaBuilder) SetKeywords(keywords string) *PageMetaBuilder {
	b.meta.Keywords = keywords
	return b
}

func (b *PageMetaBuilder) SetCanonicalURL(url string) *PageMetaBuilder {
	b.meta.CanonicalURL = url
	if b.meta.OG.URL == "" {
		b.meta.OG.URL = url
	}
	return b
}

// 页面类型设置
func (b *PageMetaBuilder) SetPageType(pageType string) *PageMetaBuilder {
	b.meta.OG.Type = pageType
	return b
}

// 文章特定设置
func (b *PageMetaBuilder) SetArticle(title, desc, author string, categories []string, publishTime, modifyTime *time.Time) *PageMetaBuilder {
	// 应用站点默认配置
	if title != "" && !strings.HasSuffix(title, b.siteConfig.SiteName) {
		title = fmt.Sprintf("%s - %s", title, b.siteConfig.SiteName)
	}
	b.SetTitle(title)
	b.SetDescription(desc)
	b.SetPageType("article")

	b.meta.Author = author

	// 设置 OpenGraph 文章信息
	b.meta.OG.ArticleAuthor = author
	b.meta.OG.ArticlePublishedTime = publishTime
	b.meta.OG.ArticleModifiedTime = modifyTime
	b.meta.OG.ArticleTag = categories

	if len(categories) > 0 {
		b.meta.OG.ArticleSection = categories[0]
		b.meta.Keywords = strings.Join(categories, ",")
	}

	return b
}

// 用户页面设置
func (b *PageMetaBuilder) SetUserProfile(username, bio string) *PageMetaBuilder {
	b.SetTitle(fmt.Sprintf("%s 的个人主页 - %s", username, b.siteConfig.SiteName))
	b.SetDescription(bio)
	b.SetPageType("profile")
	return b
}

// 图片设置
func (b *PageMetaBuilder) SetImage(imageURL, imageAlt string) *PageMetaBuilder {
	b.meta.OG.Image = imageURL
	b.meta.OG.ImageAlt = imageAlt
	b.meta.Twitter.Image = imageURL
	b.meta.Twitter.ImageAlt = imageAlt
	return b
}

// 结构化数据设置
func (b *PageMetaBuilder) SetSchemaOrg(jsonLD template.JS) *PageMetaBuilder {
	b.meta.SchemaOrgJSON = jsonLD
	return b
}

// 构建最终的 PageMeta
func (b *PageMetaBuilder) Build() *PageMeta {

	if b.meta.Description == "" {
		b.meta.Description = b.siteConfig.SiteDescription
	}

	if b.meta.Keywords == "" {
		b.meta.Keywords = b.siteConfig.SiteKeywords
	}
	return b.meta
}
