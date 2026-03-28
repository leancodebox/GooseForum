package pageConfig

import (
	"time"
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
	Version          = `version`
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
	Links []LinkItem `json:"links,omitempty"`
}

type FooterItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PItem struct {
	Content string `json:"content"`
}

// 赞助商配置

type UserSponsor struct {
	UserId    uint64 `json:"userId"`
	Amount    int    `json:"amount"` // 单位：分
	Link      string `json:"link"`
	Message   string `json:"message"`
	AvatarUrl string `json:"avatarUrl"`
	Name      string `json:"name"`
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
	Sponsors Sponsors      `json:"sponsors"`
	Users    []UserSponsor `json:"users"`
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
	ExternalLinks   string     `json:"externalLinks,omitempty"`
	FooterInfo      FooterInfo `json:"footerInfo,omitempty"`
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
	Enabled bool   `json:"enabled"`        // 是否启用公告
	Title   string `json:"title"`          // 公告标题
	Content string `json:"content"`        // 公告内容
	Link    string `json:"link,omitempty"` // 公告链接
}

type SecurityAndRegistration struct {
	EnableSignup            bool `json:"enableSignup"`
	EnableEmailVerification bool `json:"enableEmailVerification"`
	MustApproveUsers        bool `json:"mustApproveUsers"`
	MinPasswordLength       int  `json:"minPasswordLength"`
	InviteOnly              bool `json:"inviteOnly"`
	Restrictions            struct {
		AllowedDomains        []string `json:"allowedDomains"`
		BlockedDomains        []string `json:"blockedDomains"`
		MaxRegistrationsPerIp int      `json:"maxRegistrationsPerIp"`
	} `json:"restrictions"`
}

type PostingContent struct {
	TextControl struct {
		MinPostLength       int  `json:"minPostLength"`
		MaxPostLength       int  `json:"maxPostLength"`
		MinTitleLength      int  `json:"minTitleLength"`
		MaxTitleLength      int  `json:"maxTitleLength"`
		AllowUppercasePosts bool `json:"allowUppercasePosts"`
	} `json:"textControl"`
	UploadControl struct {
		AllowAttachments      bool     `json:"allowAttachments"`
		AuthorizedExtensions  []string `json:"authorizedExtensions"`
		MaxAttachmentSizeKb   int      `json:"maxAttachmentSizeKb"`
		MaxAttachmentsPerPost int      `json:"maxAttachmentsPerPost"`
	} `json:"uploadControl"`
	EditControl struct {
		EditingGracePeriod      int  `json:"editingGracePeriod"`
		PostEditTimeLimit       int  `json:"postEditTimeLimit"`
		AllowUsersToDeletePosts bool `json:"allowUsersToDeletePosts"`
	} `json:"editControl"`
}
