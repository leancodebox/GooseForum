package pageConfig

import (
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
)

const tableName = "page_config"
const pid = "id"
const filedPageType = "page_type"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                              // 主键
	PageType  string    `gorm:"column:page_type;uniqueIndex;type:varchar(128);not null;default:'';" json:"pageType"` // 页面类型
	Config    string    `gorm:"column:config;type:text;" json:"content"`                                             //
	CreatedAt time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`                  //
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

// func (itself *Entity) BeforeSave(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeCreate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterCreate(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeUpdate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterUpdate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterSave(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeDelete(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterDelete(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterFind(tx *gorm.DB) (err error) {}

func (itself *Entity) TableName() string {
	return tableName
}

const (
	FriendShipLinks  = `friendShipLinks`
	SponsorsPage     = `sponsors`
	SiteSettings     = `siteSettings`
	EmailSettings    = `emailSetting`
	Announcement     = `announcement`
	SecuritySettings = `securitySettings`
	PostingSettings  = `postingSettings`
	SiteTheme        = `siteTheme`
	Version          = `version`
	Migration        = `migration`
)

type LinkItem struct {
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Url     string `json:"url"`
	LogoUrl string `json:"logoUrl"`
	Status  int    `json:"status"`
}

type FriendLinksGroup struct {
	Name  string     `json:"name,omitempty"`
	Emoji string     `json:"emoji,omitempty"`
	Color string     `json:"color,omitempty"`
	Links []LinkItem `json:"links"`
}

type FooterItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PItem struct {
	Content string `json:"content"`
}

type SponsorItem struct {
	Link      string `json:"link"`
	Message   string `json:"message"`
	AvatarUrl string `json:"avatarUrl"`
	Name      string `json:"name"`
}

type Sponsors struct {
	Level0 []SponsorItem `json:"level0"`
	Level1 []SponsorItem `json:"level1"`
	Level2 []SponsorItem `json:"level2"`
	Level3 []SponsorItem `json:"level3"`
}

type SponsorsConfig struct {
	Sponsors Sponsors          `json:"sponsors"`
	Content  SponsorsPageIntro `json:"content"`
	Contact  SponsorsContact   `json:"contact"`
	Rules    []SponsorsRule    `json:"rules"`
}

type SponsorsPageIntro struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type SponsorsContact struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ButtonText  string `json:"buttonText"`
	ButtonLink  string `json:"buttonLink"`
}

type SponsorsRule struct {
	Content string `json:"content"`
}

// SiteSettingsConfig 站点设置配置
type SiteSettingsConfig struct {
	// 站点基本信息
	SiteName        string     `json:"siteName"`
	SiteLogo        string     `json:"siteLogo"`
	SiteDescription string     `json:"siteDescription"`
	SiteKeywords    string     `json:"siteKeywords"`
	SiteUrl         string     `json:"siteUrl"`
	SiteEmail       string     `json:"siteEmail"`
	ExternalLinks   string     `json:"externalLinks"`
	FooterInfo      FooterInfo `json:"footerInfo"`
	// 品牌标识类型: default(默认样式), text(自定义文字), image(图片)
	BrandType  string `json:"brandType"`
	BrandText  string `json:"brandText"`
	BrandImage string `json:"brandImage"`
}

type FooterInfo struct {
	Primary []PItem      `json:"primary"`
	List    []FooterItem `json:"list"`
}

// MailSettingsConfig 邮件设置配置
type MailSettingsConfig struct {
	// SMTP服务器设置
	EnableMail   bool   `json:"enableMail"`
	SmtpHost     string `json:"smtpHost"`
	SmtpPort     int    `json:"smtpPort"`
	UseSSL       bool   `json:"useSSL"`
	SmtpUsername string `json:"smtpUsername"`
	SmtpPassword string `json:"smtpPassword"`
	FromName     string `json:"fromName"`
	FromEmail    string `json:"fromEmail"`
}

// AnnouncementConfig 公告设置配置
type AnnouncementConfig struct {
	Enabled     bool   `json:"enabled"` // 是否启用公告
	Content     string `json:"content"` // 公告内容
	HtmlContent string `json:"-"`       // 预渲染后的 HTML，仅服务端使用
}

func (itself *AnnouncementConfig) PrepareHTML() {
	if itself == nil || itself.HtmlContent != "" || itself.Content == "" {
		return
	}
	itself.HtmlContent = markdown2html.MarkdownToHTML(itself.Content)
}

func (itself AnnouncementConfig) GetHtmlContent() string {
	if itself.HtmlContent != "" || itself.Content == "" {
		return itself.HtmlContent
	}
	return markdown2html.MarkdownToHTML(itself.Content)
}

type SecurityAndRegistration struct {
	EnableSignup            bool     `json:"enableSignup"`
	EnableEmailVerification bool     `json:"enableEmailVerification"`
	AllowedDomains          []string `json:"allowedDomains"`
}

type PostingContent struct {
	TextControl struct {
		MinPostLength              int `json:"minPostLength"`
		MaxPostLength              int `json:"maxPostLength"`
		MinTitleLength             int `json:"minTitleLength"`
		MaxTitleLength             int `json:"maxTitleLength"`
		NewUserPostCooldownMinutes int `json:"newUserPostCooldownMinutes"`
	} `json:"textControl"`
	UploadControl struct {
		AllowAttachments             bool     `json:"allowAttachments"`
		AuthorizedExtensions         []string `json:"authorizedExtensions"`
		MaxAttachmentSizeKb          int      `json:"maxAttachmentSizeKb"`
		MaxDailyUploadsPerUser       int      `json:"maxDailyUploadsPerUser"`
		NewUserUploadCooldownMinutes int      `json:"newUserUploadCooldownMinutes"`
	} `json:"uploadControl"`
}

type SiteThemeConfig struct {
	Version     int                   `json:"version"`
	Enabled     bool                  `json:"enabled"`
	Themes      []SiteThemeDefinition `json:"themes"`
	Prepublish  *SiteThemePrepublish  `json:"prepublish,omitempty"`
	PublishedAt string                `json:"publishedAt,omitempty"`
}

type SiteThemeTokens struct {
	ColorBase100          string `json:"color-base-100"`
	ColorBase200          string `json:"color-base-200"`
	ColorBase300          string `json:"color-base-300"`
	ColorBaseContent      string `json:"color-base-content"`
	ColorIconMuted        string `json:"color-icon-muted"`
	ColorLine             string `json:"color-line"`
	ColorPrimary          string `json:"color-primary"`
	ColorPrimaryContent   string `json:"color-primary-content"`
	ColorSecondary        string `json:"color-secondary"`
	ColorSecondaryContent string `json:"color-secondary-content"`
	ColorAccent           string `json:"color-accent"`
	ColorAccentContent    string `json:"color-accent-content"`
	ColorNeutral          string `json:"color-neutral"`
	ColorNeutralContent   string `json:"color-neutral-content"`
	ColorInfo             string `json:"color-info"`
	ColorInfoContent      string `json:"color-info-content"`
	ColorSuccess          string `json:"color-success"`
	ColorSuccessContent   string `json:"color-success-content"`
	ColorWarning          string `json:"color-warning"`
	ColorWarningContent   string `json:"color-warning-content"`
	ColorError            string `json:"color-error"`
	ColorErrorContent     string `json:"color-error-content"`
	RadiusSelector        string `json:"radius-selector"`
	RadiusField           string `json:"radius-field"`
	RadiusBox             string `json:"radius-box"`
	SizeSelector          string `json:"size-selector"`
	SizeField             string `json:"size-field"`
	Border                string `json:"border"`
	Depth                 string `json:"depth"`
}

func (tokens *SiteThemeTokens) NormalizeFrom(defaults SiteThemeTokens) {
	tokens.ColorBase100 = normalizeSiteThemeTokenValue(tokens.ColorBase100, defaults.ColorBase100)
	tokens.ColorBase200 = normalizeSiteThemeTokenValue(tokens.ColorBase200, defaults.ColorBase200)
	tokens.ColorBase300 = normalizeSiteThemeTokenValue(tokens.ColorBase300, defaults.ColorBase300)
	tokens.ColorBaseContent = normalizeSiteThemeTokenValue(tokens.ColorBaseContent, defaults.ColorBaseContent)
	tokens.ColorIconMuted = normalizeSiteThemeTokenValue(tokens.ColorIconMuted, defaults.ColorIconMuted)
	tokens.ColorLine = normalizeSiteThemeTokenValue(tokens.ColorLine, defaults.ColorLine)
	tokens.ColorPrimary = normalizeSiteThemeTokenValue(tokens.ColorPrimary, defaults.ColorPrimary)
	tokens.ColorPrimaryContent = normalizeSiteThemeTokenValue(tokens.ColorPrimaryContent, defaults.ColorPrimaryContent)
	tokens.ColorSecondary = normalizeSiteThemeTokenValue(tokens.ColorSecondary, defaults.ColorSecondary)
	tokens.ColorSecondaryContent = normalizeSiteThemeTokenValue(tokens.ColorSecondaryContent, defaults.ColorSecondaryContent)
	tokens.ColorAccent = normalizeSiteThemeTokenValue(tokens.ColorAccent, defaults.ColorAccent)
	tokens.ColorAccentContent = normalizeSiteThemeTokenValue(tokens.ColorAccentContent, defaults.ColorAccentContent)
	tokens.ColorNeutral = normalizeSiteThemeTokenValue(tokens.ColorNeutral, defaults.ColorNeutral)
	tokens.ColorNeutralContent = normalizeSiteThemeTokenValue(tokens.ColorNeutralContent, defaults.ColorNeutralContent)
	tokens.ColorInfo = normalizeSiteThemeTokenValue(tokens.ColorInfo, defaults.ColorInfo)
	tokens.ColorInfoContent = normalizeSiteThemeTokenValue(tokens.ColorInfoContent, defaults.ColorInfoContent)
	tokens.ColorSuccess = normalizeSiteThemeTokenValue(tokens.ColorSuccess, defaults.ColorSuccess)
	tokens.ColorSuccessContent = normalizeSiteThemeTokenValue(tokens.ColorSuccessContent, defaults.ColorSuccessContent)
	tokens.ColorWarning = normalizeSiteThemeTokenValue(tokens.ColorWarning, defaults.ColorWarning)
	tokens.ColorWarningContent = normalizeSiteThemeTokenValue(tokens.ColorWarningContent, defaults.ColorWarningContent)
	tokens.ColorError = normalizeSiteThemeTokenValue(tokens.ColorError, defaults.ColorError)
	tokens.ColorErrorContent = normalizeSiteThemeTokenValue(tokens.ColorErrorContent, defaults.ColorErrorContent)
	tokens.RadiusSelector = normalizeSiteThemeTokenValue(tokens.RadiusSelector, defaults.RadiusSelector)
	tokens.RadiusField = normalizeLegacyRadiusField(normalizeSiteThemeTokenValue(tokens.RadiusField, defaults.RadiusField))
	tokens.RadiusBox = normalizeSiteThemeTokenValue(tokens.RadiusBox, defaults.RadiusBox)
	tokens.SizeSelector = normalizeSiteThemeTokenValue(tokens.SizeSelector, defaults.SizeSelector)
	tokens.SizeField = normalizeSiteThemeTokenValue(tokens.SizeField, defaults.SizeField)
	tokens.Border = normalizeSiteThemeTokenValue(tokens.Border, defaults.Border)
	tokens.Depth = normalizeSiteThemeTokenValue(tokens.Depth, defaults.Depth)
}

func (tokens SiteThemeTokens) BaseColor() string {
	return sanitizeSiteThemeTokenValue(tokens.ColorBase100)
}

func (tokens SiteThemeTokens) AppendCSSVariables(sb *strings.Builder) {
	appendSiteThemeCSSVar(sb, "color-base-100", tokens.ColorBase100)
	appendSiteThemeCSSVar(sb, "color-base-200", tokens.ColorBase200)
	appendSiteThemeCSSVar(sb, "color-base-300", tokens.ColorBase300)
	appendSiteThemeCSSVar(sb, "color-base-content", tokens.ColorBaseContent)
	appendSiteThemeCSSVar(sb, "color-icon-muted", tokens.ColorIconMuted)
	appendSiteThemeCSSVar(sb, "color-line", tokens.ColorLine)
	appendSiteThemeCSSVar(sb, "color-primary", tokens.ColorPrimary)
	appendSiteThemeCSSVar(sb, "color-primary-content", tokens.ColorPrimaryContent)
	appendSiteThemeCSSVar(sb, "color-secondary", tokens.ColorSecondary)
	appendSiteThemeCSSVar(sb, "color-secondary-content", tokens.ColorSecondaryContent)
	appendSiteThemeCSSVar(sb, "color-accent", tokens.ColorAccent)
	appendSiteThemeCSSVar(sb, "color-accent-content", tokens.ColorAccentContent)
	appendSiteThemeCSSVar(sb, "color-neutral", tokens.ColorNeutral)
	appendSiteThemeCSSVar(sb, "color-neutral-content", tokens.ColorNeutralContent)
	appendSiteThemeCSSVar(sb, "color-info", tokens.ColorInfo)
	appendSiteThemeCSSVar(sb, "color-info-content", tokens.ColorInfoContent)
	appendSiteThemeCSSVar(sb, "color-success", tokens.ColorSuccess)
	appendSiteThemeCSSVar(sb, "color-success-content", tokens.ColorSuccessContent)
	appendSiteThemeCSSVar(sb, "color-warning", tokens.ColorWarning)
	appendSiteThemeCSSVar(sb, "color-warning-content", tokens.ColorWarningContent)
	appendSiteThemeCSSVar(sb, "color-error", tokens.ColorError)
	appendSiteThemeCSSVar(sb, "color-error-content", tokens.ColorErrorContent)
	appendSiteThemeCSSVar(sb, "radius-selector", tokens.RadiusSelector)
	appendSiteThemeCSSVar(sb, "radius-field", tokens.RadiusField)
	appendSiteThemeCSSVar(sb, "radius-box", tokens.RadiusBox)
	appendSiteThemeCSSVar(sb, "size-selector", tokens.SizeSelector)
	appendSiteThemeCSSVar(sb, "size-field", tokens.SizeField)
	appendSiteThemeCSSVar(sb, "border", tokens.Border)
	appendSiteThemeCSSVar(sb, "depth", tokens.Depth)
}

func normalizeSiteThemeTokenValue(value string, defaultValue string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "{};<>") {
		return defaultValue
	}
	return value
}

func normalizeLegacyRadiusField(value string) string {
	switch strings.TrimSpace(value) {
	case "0.375rem", "6px":
		return "0.5rem"
	default:
		return value
	}
}

func sanitizeSiteThemeTokenValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "{};<>") {
		return ""
	}
	return value
}

func appendSiteThemeCSSVar(sb *strings.Builder, name string, value string) {
	value = sanitizeSiteThemeTokenValue(value)
	if value == "" {
		return
	}
	sb.WriteString("--gf-")
	sb.WriteString(name)
	sb.WriteByte(':')
	sb.WriteString(value)
	sb.WriteByte(';')
}

type SiteThemeDefinition struct {
	Name        string          `json:"name"`
	Label       string          `json:"label"`
	ColorScheme string          `json:"colorScheme"`
	Tokens      SiteThemeTokens `json:"tokens"`
}

type SiteThemePrepublish struct {
	Enabled   bool                  `json:"enabled"`
	Themes    []SiteThemeDefinition `json:"themes"`
	UpdatedAt string                `json:"updatedAt,omitempty"`
}

func FirstSiteThemeDefinition(themes []SiteThemeDefinition) SiteThemeDefinition {
	if len(themes) == 0 {
		return SiteThemeDefinition{}
	}
	return themes[0]
}

func NormalizeSiteThemeDefinitions(themes []SiteThemeDefinition, defaults []SiteThemeDefinition, fallback SiteThemeDefinition) []SiteThemeDefinition {
	if len(themes) == 0 {
		return cloneSiteThemeDefinitions(defaults)
	}
	for index := range themes {
		NormalizeSiteThemeDefinition(&themes[index], defaults, fallback)
	}
	return themes
}

func NormalizeSiteThemeDefinition(theme *SiteThemeDefinition, defaults []SiteThemeDefinition, fallback SiteThemeDefinition) {
	defaultTheme := defaultSiteThemeDefinition(theme.Name, defaults, fallback)
	if theme.Name != defaultTheme.Name {
		theme.Name = defaultTheme.Name
	}
	if theme.Label == "" {
		theme.Label = defaultTheme.Label
	}
	if !isSiteThemeColorScheme(theme.ColorScheme) {
		theme.ColorScheme = defaultTheme.ColorScheme
	}
	theme.Tokens.NormalizeFrom(defaultTheme.Tokens)
}

func cloneSiteThemeDefinitions(themes []SiteThemeDefinition) []SiteThemeDefinition {
	cloned := make([]SiteThemeDefinition, len(themes))
	copy(cloned, themes)
	return cloned
}

func defaultSiteThemeDefinition(name string, defaults []SiteThemeDefinition, fallback SiteThemeDefinition) SiteThemeDefinition {
	name = strings.TrimSpace(name)
	for _, theme := range defaults {
		if theme.Name == name {
			return theme
		}
	}
	return fallback
}

func isSiteThemeColorScheme(value string) bool {
	return value == "dark" || value == "light"
}
