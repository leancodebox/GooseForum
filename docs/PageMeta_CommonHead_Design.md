# PageMeta 和 CommonHead 设计文档

## 概述

本文档描述了 GooseForum 中 PageMeta 结构体和 common-head.gohtml 模板的设计方案，旨在提供统一、灵活且 SEO 友好的页面元数据管理机制。

## 当前状态分析

### 现有 PageMeta 结构体

```go
// app/http/controllers/viewrender/struct.go
type PageMeta struct {
    // 基础SEO
    Title        string
    Description  string
    Keywords     string
    CanonicalURL string

    // OpenGraph
    OGType        string
    OGTitle       string
    OGDescription string
    OGImage       string
    OGURL         string

    Favicon    string
    ThemeColor string

    // 结构化数据
    SchemaOrgJSON string
}
```

### 现有 common-head.gohtml 模板

```html
<!-- resource/templates/layout/common-head.gohtml -->
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<meta name="keywords" content="{{if .Keywords}}{{.Keywords}}{{else}}GooseForum{{end}}">
<meta property="og:site_name" content="GooseForum">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:type" content="{{if .OgType}}{{.OgType}}{{else}}website{{end}}">
<link rel="alternate" type="application/rss+xml" title="GooseForum RSS Feed" href="/rss.xml"/>
{{if .CanonicalHref}}
<link rel="canonical" href="{{.CanonicalHref}}">
<meta property="og:url" content="{{.CanonicalHref}}"/>{{end}}
<link rel="icon" type="image/png" href="/static/pic/icon.webp">
{{with WebPageSettings}}{{SafeHTML .ExternalLinks}}{{end}}
```

### 当前问题

1. **数据传递不统一**：各控制器直接在 map[string]any 中传递零散的 SEO 字段
2. **缺乏标准化**：没有统一的 PageMeta 构建机制
3. **模板耦合度高**：common-head.gohtml 直接访问多个不同的字段名
4. **扩展性差**：添加新的 SEO 字段需要修改多个地方
5. **缺乏默认值机制**：没有全局的默认 SEO 配置

## 重构设计方案

### 1. 增强的 PageMeta 结构体

```go
// app/http/controllers/viewrender/struct.go
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
    
    // 页面类型和状态
    PageType   string `json:"page_type,omitempty"` // article, website, profile, etc.
    Language   string `json:"language,omitempty"`
    Robots     string `json:"robots,omitempty"`    // index,follow / noindex,nofollow
    
    // 时间信息
    PublishedTime *time.Time `json:"published_time,omitempty"`
    ModifiedTime  *time.Time `json:"modified_time,omitempty"`
    
    // 结构化数据
    SchemaOrgJSON string `json:"schema_org_json,omitempty"`
    
    // 页面特定配置
    Favicon    string `json:"favicon,omitempty"`
    ThemeColor string `json:"theme_color,omitempty"`
    
    // 扩展字段
    CustomMeta map[string]string `json:"custom_meta,omitempty"`
}

type OpenGraphMeta struct {
    Type        string `json:"type,omitempty"`        // website, article, profile
    Title       string `json:"title,omitempty"`
    Description string `json:"description,omitempty"`
    Image       string `json:"image,omitempty"`
    ImageAlt    string `json:"image_alt,omitempty"`
    URL         string `json:"url,omitempty"`
    SiteName    string `json:"site_name,omitempty"`
    Locale      string `json:"locale,omitempty"`
    
    // 文章特定字段
    ArticleAuthor      string     `json:"article_author,omitempty"`
    ArticleSection     string     `json:"article_section,omitempty"`
    ArticleTag         []string   `json:"article_tag,omitempty"`
    ArticlePublishedTime *time.Time `json:"article_published_time,omitempty"`
    ArticleModifiedTime  *time.Time `json:"article_modified_time,omitempty"`
}

type TwitterCardMeta struct {
    Card        string `json:"card,omitempty"`        // summary, summary_large_image
    Site        string `json:"site,omitempty"`        // @username
    Creator     string `json:"creator,omitempty"`     // @username
    Title       string `json:"title,omitempty"`
    Description string `json:"description,omitempty"`
    Image       string `json:"image,omitempty"`
    ImageAlt    string `json:"image_alt,omitempty"`
}
```

### 2. PageMeta 构建器

```go
// app/http/controllers/viewrender/builder.go
type PageMetaBuilder struct {
    meta     *PageMeta
    siteConfig *pageConfig.SiteSettingsConfig
}

func NewPageMetaBuilder() *PageMetaBuilder {
    siteConfig := hotdataserve.GetSiteSettingsConfigCache()
    
    return &PageMetaBuilder{
        meta: &PageMeta{
            // 设置默认值
            Language:   "zh-CN",
            Robots:     "index,follow",
            Favicon:    "/static/pic/icon.webp",
            ThemeColor: "#1976d2",
            OG: OpenGraphMeta{
                Type:     "website",
                SiteName: siteConfig.SiteName,
                Locale:   "zh_CN",
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
    b.meta.PageType = pageType
    b.meta.OG.Type = pageType
    return b
}

// 文章特定设置
func (b *PageMetaBuilder) SetArticle(title, desc, author string, categories []string, publishTime, modifyTime *time.Time) *PageMetaBuilder {
    b.SetTitle(title)
    b.SetDescription(desc)
    b.SetPageType("article")
    
    b.meta.Author = author
    b.meta.PublishedTime = publishTime
    b.meta.ModifiedTime = modifyTime
    
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
func (b *PageMetaBuilder) SetSchemaOrg(jsonLD string) *PageMetaBuilder {
    b.meta.SchemaOrgJSON = jsonLD
    return b
}

// 自定义 meta 标签
func (b *PageMetaBuilder) AddCustomMeta(name, content string) *PageMetaBuilder {
    if b.meta.CustomMeta == nil {
        b.meta.CustomMeta = make(map[string]string)
    }
    b.meta.CustomMeta[name] = content
    return b
}

// 构建最终的 PageMeta
func (b *PageMetaBuilder) Build() *PageMeta {
    // 应用站点默认配置
    if b.meta.Title != "" && !strings.Contains(b.meta.Title, b.siteConfig.SiteName) {
        b.meta.Title = fmt.Sprintf("%s - %s", b.meta.Title, b.siteConfig.SiteName)
    }
    
    if b.meta.Description == "" {
        b.meta.Description = b.siteConfig.SiteDescription
    }
    
    if b.meta.Keywords == "" {
        b.meta.Keywords = b.siteConfig.SiteKeywords
    }
    
    return b.meta
}
```

### 3. 重构后的 common-head.gohtml

```html
<!-- resource/templates/layout/common-head.gohtml -->
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">

{{with .PageMeta}}
<!-- 基础 SEO 标签 -->
<title>{{.Title}}</title>
{{if .Description}}<meta name="description" content="{{.Description}}">{{end}}
{{if .Keywords}}<meta name="keywords" content="{{.Keywords}}">{{end}}
{{if .Author}}<meta name="author" content="{{.Author}}">{{end}}
{{if .Language}}<meta name="language" content="{{.Language}}">{{end}}
{{if .Robots}}<meta name="robots" content="{{.Robots}}">{{end}}

<!-- URL 相关 -->
{{if .CanonicalURL}}<link rel="canonical" href="{{.CanonicalURL}}">{{end}}
{{if .AlternateURL}}<link rel="alternate" href="{{.AlternateURL}}">{{end}}

<!-- Open Graph 标签 -->
{{with .OG}}
{{if .Type}}<meta property="og:type" content="{{.Type}}">{{end}}
{{if .Title}}<meta property="og:title" content="{{.Title}}">{{end}}
{{if .Description}}<meta property="og:description" content="{{.Description}}">{{end}}
{{if .Image}}<meta property="og:image" content="{{.Image}}">{{end}}
{{if .ImageAlt}}<meta property="og:image:alt" content="{{.ImageAlt}}">{{end}}
{{if .URL}}<meta property="og:url" content="{{.URL}}">{{end}}
{{if .SiteName}}<meta property="og:site_name" content="{{.SiteName}}">{{end}}
{{if .Locale}}<meta property="og:locale" content="{{.Locale}}">{{end}}

<!-- 文章特定的 Open Graph 标签 -->
{{if .ArticleAuthor}}<meta property="article:author" content="{{.ArticleAuthor}}">{{end}}
{{if .ArticleSection}}<meta property="article:section" content="{{.ArticleSection}}">{{end}}
{{range .ArticleTag}}<meta property="article:tag" content="{{.}}">{{end}}
{{if .ArticlePublishedTime}}<meta property="article:published_time" content="{{.ArticlePublishedTime.Format "2006-01-02T15:04:05Z07:00"}}">{{end}}
{{if .ArticleModifiedTime}}<meta property="article:modified_time" content="{{.ArticleModifiedTime.Format "2006-01-02T15:04:05Z07:00"}}">{{end}}
{{end}}

<!-- Twitter Card 标签 -->
{{with .Twitter}}
{{if .Card}}<meta name="twitter:card" content="{{.Card}}">{{end}}
{{if .Site}}<meta name="twitter:site" content="{{.Site}}">{{end}}
{{if .Creator}}<meta name="twitter:creator" content="{{.Creator}}">{{end}}
{{if .Title}}<meta name="twitter:title" content="{{.Title}}">{{end}}
{{if .Description}}<meta name="twitter:description" content="{{.Description}}">{{end}}
{{if .Image}}<meta name="twitter:image" content="{{.Image}}">{{end}}
{{if .ImageAlt}}<meta name="twitter:image:alt" content="{{.ImageAlt}}">{{end}}
{{end}}

<!-- 时间信息 -->
{{if .PublishedTime}}<meta name="article:published_time" content="{{.PublishedTime.Format "2006-01-02T15:04:05Z07:00"}}">{{end}}
{{if .ModifiedTime}}<meta name="article:modified_time" content="{{.ModifiedTime.Format "2006-01-02T15:04:05Z07:00"}}">{{end}}

<!-- 页面配置 -->
{{if .Favicon}}<link rel="icon" type="image/png" href="{{.Favicon}}">{{end}}
{{if .ThemeColor}}<meta name="theme-color" content="{{.ThemeColor}}">{{end}}

<!-- 自定义 meta 标签 -->
{{range $name, $content := .CustomMeta}}
<meta name="{{$name}}" content="{{$content}}">
{{end}}

<!-- 结构化数据 -->
{{if .SchemaOrgJSON}}
<script type="application/ld+json">
{{.SchemaOrgJSON | SafeHTML}}
</script>
{{end}}
{{end}}

<!-- RSS 订阅 -->
<link rel="alternate" type="application/rss+xml" title="GooseForum RSS Feed" href="/rss.xml"/>

<!-- 外部链接配置 -->
{{with WebPageSettings}}{{SafeHTML .ExternalLinks}}{{end}}
```

### 4. 控制器使用示例

#### 首页控制器
```go
func Home(c *gin.Context) {
    pageMeta := viewrender.NewPageMetaBuilder().
        SetTitle("GooseForum - 自由漫谈的江湖茶馆").
        SetDescription("GooseForum's home").
        SetCanonicalURL(buildCanonicalHref(c)).
        Build()
    
    viewrender.Render(c, "index.gohtml", map[string]any{
        "PageMeta":            pageMeta,
        "User":                GetLoginUser(c),
        "ArticleCategoryList": articleCategoryLabel(),
        "LatestArticles":      latestArticles,
        "Stats":               GetSiteStatisticsData(),
        "RecommendedArticles": getRecommendedArticles(),
        "GooseForumInfo":      GetGooseForumInfo(),
    })
}
```

#### 文章详情控制器
```go
func PostDetail(c *gin.Context) {
    // ... 获取文章数据逻辑 ...
    
    pageMeta := viewrender.NewPageMetaBuilder().
        SetArticle(
            entity.Title,
            entity.Description,
            author,
            articleCategory,
            &entity.CreatedAt,
            &entity.UpdatedAt,
        ).
        SetCanonicalURL(buildCanonicalHref(c)).
        SetSchemaOrg(string(generateArticleJSONLD(c, entity, author))).
        Build()
    
    viewrender.Render(c, "detail.gohtml", map[string]any{
        "PageMeta":             pageMeta,
        "ArticleId":            id,
        "AuthorId":             authorId,
        "ArticleTitle":         entity.Title,
        "ArticleContent":       template.HTML(entity.RenderedHTML),
        // ... 其他数据 ...
    })
}
```

#### 用户页面控制器
```go
func User(c *gin.Context) {
    // ... 获取用户数据逻辑 ...
    
    pageMeta := viewrender.NewPageMetaBuilder().
        SetUserProfile(userInfo.Username, userInfo.Bio).
        SetCanonicalURL(buildCanonicalHref(c)).
        Build()
    
    viewrender.Render(c, "user.gohtml", map[string]any{
        "PageMeta": pageMeta,
        // ... 其他数据 ...
    })
}
```

## 实施计划

### 阶段 1：基础结构重构
1. 创建新的 PageMeta 结构体和相关类型
2. 实现 PageMetaBuilder
3. 更新 common-head.gohtml 模板

### 阶段 2：控制器迁移
1. 逐步迁移各个控制器使用新的 PageMeta 构建器
2. 移除旧的零散 SEO 字段传递
3. 测试各页面的 SEO 标签输出

### 阶段 3：功能增强
1. 添加更多结构化数据支持
2. 实现动态 SEO 配置
3. 添加 SEO 预览和验证工具

## 优势

1. **统一性**：所有页面使用统一的 PageMeta 结构
2. **可维护性**：集中管理 SEO 相关逻辑
3. **扩展性**：易于添加新的 SEO 字段和功能
4. **类型安全**：使用结构体替代 map[string]any
5. **默认值支持**：自动应用站点级别的默认配置
6. **SEO 友好**：支持完整的 Open Graph、Twitter Card 和结构化数据

## 注意事项

1. 迁移过程中需要保持向后兼容
2. 需要充分测试各种页面类型的 SEO 输出
3. 考虑性能影响，避免过度的字符串拼接
4. 确保所有 SEO 标签都正确转义，防止 XSS 攻击
