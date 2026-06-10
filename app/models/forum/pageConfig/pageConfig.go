package pageConfig

import (
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
	FooterInfo      FooterInfo `json:"footerInfo,omitempty"`
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
	Draft       *SiteThemeSnapshot    `json:"draft,omitempty"`
	History     []SiteThemeSnapshot   `json:"history,omitempty"`
	PublishedAt string                `json:"publishedAt,omitempty"`
}

type SiteThemeTokenKey string

const (
	SiteThemeTokenColorBase100          SiteThemeTokenKey = "color-base-100"
	SiteThemeTokenColorBase200          SiteThemeTokenKey = "color-base-200"
	SiteThemeTokenColorBase300          SiteThemeTokenKey = "color-base-300"
	SiteThemeTokenColorBaseContent      SiteThemeTokenKey = "color-base-content"
	SiteThemeTokenColorIconMuted        SiteThemeTokenKey = "color-icon-muted"
	SiteThemeTokenColorLine             SiteThemeTokenKey = "color-line"
	SiteThemeTokenColorPrimary          SiteThemeTokenKey = "color-primary"
	SiteThemeTokenColorPrimaryContent   SiteThemeTokenKey = "color-primary-content"
	SiteThemeTokenColorSecondary        SiteThemeTokenKey = "color-secondary"
	SiteThemeTokenColorSecondaryContent SiteThemeTokenKey = "color-secondary-content"
	SiteThemeTokenColorAccent           SiteThemeTokenKey = "color-accent"
	SiteThemeTokenColorAccentContent    SiteThemeTokenKey = "color-accent-content"
	SiteThemeTokenColorNeutral          SiteThemeTokenKey = "color-neutral"
	SiteThemeTokenColorNeutralContent   SiteThemeTokenKey = "color-neutral-content"
	SiteThemeTokenColorInfo             SiteThemeTokenKey = "color-info"
	SiteThemeTokenColorInfoContent      SiteThemeTokenKey = "color-info-content"
	SiteThemeTokenColorSuccess          SiteThemeTokenKey = "color-success"
	SiteThemeTokenColorSuccessContent   SiteThemeTokenKey = "color-success-content"
	SiteThemeTokenColorWarning          SiteThemeTokenKey = "color-warning"
	SiteThemeTokenColorWarningContent   SiteThemeTokenKey = "color-warning-content"
	SiteThemeTokenColorError            SiteThemeTokenKey = "color-error"
	SiteThemeTokenColorErrorContent     SiteThemeTokenKey = "color-error-content"
	SiteThemeTokenRadiusSelector        SiteThemeTokenKey = "radius-selector"
	SiteThemeTokenRadiusField           SiteThemeTokenKey = "radius-field"
	SiteThemeTokenRadiusBox             SiteThemeTokenKey = "radius-box"
	SiteThemeTokenSizeSelector          SiteThemeTokenKey = "size-selector"
	SiteThemeTokenSizeField             SiteThemeTokenKey = "size-field"
	SiteThemeTokenBorder                SiteThemeTokenKey = "border"
	SiteThemeTokenDepth                 SiteThemeTokenKey = "depth"
	SiteThemeTokenNoise                 SiteThemeTokenKey = "noise"
)

var siteThemeTokenKeys = []SiteThemeTokenKey{
	SiteThemeTokenColorBase100,
	SiteThemeTokenColorBase200,
	SiteThemeTokenColorBase300,
	SiteThemeTokenColorBaseContent,
	SiteThemeTokenColorIconMuted,
	SiteThemeTokenColorLine,
	SiteThemeTokenColorPrimary,
	SiteThemeTokenColorPrimaryContent,
	SiteThemeTokenColorSecondary,
	SiteThemeTokenColorSecondaryContent,
	SiteThemeTokenColorAccent,
	SiteThemeTokenColorAccentContent,
	SiteThemeTokenColorNeutral,
	SiteThemeTokenColorNeutralContent,
	SiteThemeTokenColorInfo,
	SiteThemeTokenColorInfoContent,
	SiteThemeTokenColorSuccess,
	SiteThemeTokenColorSuccessContent,
	SiteThemeTokenColorWarning,
	SiteThemeTokenColorWarningContent,
	SiteThemeTokenColorError,
	SiteThemeTokenColorErrorContent,
	SiteThemeTokenRadiusSelector,
	SiteThemeTokenRadiusField,
	SiteThemeTokenRadiusBox,
	SiteThemeTokenSizeSelector,
	SiteThemeTokenSizeField,
	SiteThemeTokenBorder,
	SiteThemeTokenDepth,
	SiteThemeTokenNoise,
}

func SiteThemeTokenKeys() []SiteThemeTokenKey {
	return append([]SiteThemeTokenKey(nil), siteThemeTokenKeys...)
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
	Noise                 string `json:"noise"`
}

func (tokens SiteThemeTokens) Get(key SiteThemeTokenKey) string {
	switch key {
	case SiteThemeTokenColorBase100:
		return tokens.ColorBase100
	case SiteThemeTokenColorBase200:
		return tokens.ColorBase200
	case SiteThemeTokenColorBase300:
		return tokens.ColorBase300
	case SiteThemeTokenColorBaseContent:
		return tokens.ColorBaseContent
	case SiteThemeTokenColorIconMuted:
		return tokens.ColorIconMuted
	case SiteThemeTokenColorLine:
		return tokens.ColorLine
	case SiteThemeTokenColorPrimary:
		return tokens.ColorPrimary
	case SiteThemeTokenColorPrimaryContent:
		return tokens.ColorPrimaryContent
	case SiteThemeTokenColorSecondary:
		return tokens.ColorSecondary
	case SiteThemeTokenColorSecondaryContent:
		return tokens.ColorSecondaryContent
	case SiteThemeTokenColorAccent:
		return tokens.ColorAccent
	case SiteThemeTokenColorAccentContent:
		return tokens.ColorAccentContent
	case SiteThemeTokenColorNeutral:
		return tokens.ColorNeutral
	case SiteThemeTokenColorNeutralContent:
		return tokens.ColorNeutralContent
	case SiteThemeTokenColorInfo:
		return tokens.ColorInfo
	case SiteThemeTokenColorInfoContent:
		return tokens.ColorInfoContent
	case SiteThemeTokenColorSuccess:
		return tokens.ColorSuccess
	case SiteThemeTokenColorSuccessContent:
		return tokens.ColorSuccessContent
	case SiteThemeTokenColorWarning:
		return tokens.ColorWarning
	case SiteThemeTokenColorWarningContent:
		return tokens.ColorWarningContent
	case SiteThemeTokenColorError:
		return tokens.ColorError
	case SiteThemeTokenColorErrorContent:
		return tokens.ColorErrorContent
	case SiteThemeTokenRadiusSelector:
		return tokens.RadiusSelector
	case SiteThemeTokenRadiusField:
		return tokens.RadiusField
	case SiteThemeTokenRadiusBox:
		return tokens.RadiusBox
	case SiteThemeTokenSizeSelector:
		return tokens.SizeSelector
	case SiteThemeTokenSizeField:
		return tokens.SizeField
	case SiteThemeTokenBorder:
		return tokens.Border
	case SiteThemeTokenDepth:
		return tokens.Depth
	case SiteThemeTokenNoise:
		return tokens.Noise
	default:
		return ""
	}
}

func (tokens *SiteThemeTokens) Set(key SiteThemeTokenKey, value string) {
	switch key {
	case SiteThemeTokenColorBase100:
		tokens.ColorBase100 = value
	case SiteThemeTokenColorBase200:
		tokens.ColorBase200 = value
	case SiteThemeTokenColorBase300:
		tokens.ColorBase300 = value
	case SiteThemeTokenColorBaseContent:
		tokens.ColorBaseContent = value
	case SiteThemeTokenColorIconMuted:
		tokens.ColorIconMuted = value
	case SiteThemeTokenColorLine:
		tokens.ColorLine = value
	case SiteThemeTokenColorPrimary:
		tokens.ColorPrimary = value
	case SiteThemeTokenColorPrimaryContent:
		tokens.ColorPrimaryContent = value
	case SiteThemeTokenColorSecondary:
		tokens.ColorSecondary = value
	case SiteThemeTokenColorSecondaryContent:
		tokens.ColorSecondaryContent = value
	case SiteThemeTokenColorAccent:
		tokens.ColorAccent = value
	case SiteThemeTokenColorAccentContent:
		tokens.ColorAccentContent = value
	case SiteThemeTokenColorNeutral:
		tokens.ColorNeutral = value
	case SiteThemeTokenColorNeutralContent:
		tokens.ColorNeutralContent = value
	case SiteThemeTokenColorInfo:
		tokens.ColorInfo = value
	case SiteThemeTokenColorInfoContent:
		tokens.ColorInfoContent = value
	case SiteThemeTokenColorSuccess:
		tokens.ColorSuccess = value
	case SiteThemeTokenColorSuccessContent:
		tokens.ColorSuccessContent = value
	case SiteThemeTokenColorWarning:
		tokens.ColorWarning = value
	case SiteThemeTokenColorWarningContent:
		tokens.ColorWarningContent = value
	case SiteThemeTokenColorError:
		tokens.ColorError = value
	case SiteThemeTokenColorErrorContent:
		tokens.ColorErrorContent = value
	case SiteThemeTokenRadiusSelector:
		tokens.RadiusSelector = value
	case SiteThemeTokenRadiusField:
		tokens.RadiusField = value
	case SiteThemeTokenRadiusBox:
		tokens.RadiusBox = value
	case SiteThemeTokenSizeSelector:
		tokens.SizeSelector = value
	case SiteThemeTokenSizeField:
		tokens.SizeField = value
	case SiteThemeTokenBorder:
		tokens.Border = value
	case SiteThemeTokenDepth:
		tokens.Depth = value
	case SiteThemeTokenNoise:
		tokens.Noise = value
	}
}

type SiteThemeDefinition struct {
	Name        string          `json:"name"`
	Label       string          `json:"label"`
	ColorScheme string          `json:"colorScheme"`
	Tokens      SiteThemeTokens `json:"tokens"`
}

type SiteThemeSnapshot struct {
	Enabled   bool                  `json:"enabled"`
	Themes    []SiteThemeDefinition `json:"themes"`
	CreatedAt string                `json:"createdAt,omitempty"`
	Label     string                `json:"label,omitempty"`
}
