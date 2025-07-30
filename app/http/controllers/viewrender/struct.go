package viewrender

import (
	"html/template"
	"time"
)

type PageMeta struct {
	// 基础 SEO 信息
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	Author      string `json:"author,omitempty"`

	// URL 相关
	CanonicalURL string `json:"canonical_url,omitempty"`
	AlternateURL string `json:"alternate_url,omitempty"`

	// Open Graph 协议
	OG OpenGraphMeta `json:"og"`

	// Twitter Card
	Twitter TwitterCardMeta `json:"twitter"`

	// 结构化数据
	SchemaOrgJSON template.JS `json:"schema_org_json,omitempty"`

	// 页面特定配置
	Favicon    string `json:"favicon,omitempty"`
	ThemeColor string `json:"theme_color,omitempty"`
}

type OpenGraphMeta struct {
	Type        string `json:"type,omitempty"` // website, article, profile
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	ImageAlt    string `json:"image_alt,omitempty"`
	URL         string `json:"url,omitempty"`
	SiteName    string `json:"site_name,omitempty"`

	// 文章特定字段
	ArticleAuthor        string     `json:"article_author,omitempty"`
	ArticleSection       string     `json:"article_section,omitempty"`
	ArticleTag           []string   `json:"article_tag,omitempty"`
	ArticlePublishedTime *time.Time `json:"article_published_time,omitempty"`
	ArticleModifiedTime  *time.Time `json:"article_modified_time,omitempty"`
}

type TwitterCardMeta struct {
	Card        string `json:"card,omitempty"`    // summary, summary_large_image
	Site        string `json:"site,omitempty"`    // @username
	Creator     string `json:"creator,omitempty"` // @username
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	ImageAlt    string `json:"image_alt,omitempty"`
}
