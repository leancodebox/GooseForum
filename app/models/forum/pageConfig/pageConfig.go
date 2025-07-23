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
	FriendShipLinks = `friendShipLinks`
	WebSettings     = `webSettings`
	FooterLinks     = `footerLinks`
	SponsorsPage    = `sponsors`
	SiteSettings    = `siteSettings`
)

var PageTypeList = []string{
	FriendShipLinks,
	WebSettings,
	FooterLinks,
	SponsorsPage,
	SiteSettings,
}

type WebSettingsConfig struct {
	ExternalLinks string `json:"externalLinks,omitempty"`
	Favicon       string `json:"favicon,omitempty"`
}

type LinkItem struct {
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Url     string `json:"url"`
	LogoUrl string `json:"logoUrl"`
}

type FriendLinksGroup struct {
	Name  string     `json:"name,omitempty"`
	Links []LinkItem `json:"links,omitempty"`
}

type FooterItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type FooterGroup struct {
	Name     string       `json:"name"`
	Children []FooterItem `json:"children"`
}

type PItem struct {
	Content string `json:"content"`
}

type FooterConfig struct {
	Primary []PItem       `json:"primary"`
	List    []FooterGroup `json:"list"`
}

// 赞助商相关数据结构
type SponsorItem struct {
	Name string   `json:"name"`
	Logo string   `json:"logo"`
	Info string   `json:"info"`
	Url  string   `json:"url"`
	Tag  []string `json:"tag"`
}

type UserSponsor struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Time   string `json:"time"`
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

// 站点设置配置
type SiteSettingsConfig struct {
	// 站点基本信息
	SiteName        string `json:"siteName"`
	SiteLogo        string `json:"siteLogo"`
	SiteDescription string `json:"siteDescription"`
	SiteKeywords    string `json:"siteKeywords"`
	SiteUrl         string `json:"siteUrl"`
	
	// SEO设置
	TitleTemplate      string `json:"titleTemplate"`
	DefaultDescription string `json:"defaultDescription"`
	IcpNumber          string `json:"icpNumber"`
	
	// 其他设置
	Timezone           string `json:"timezone"`
	DefaultLanguage    string `json:"defaultLanguage"`
	MaintenanceMode    bool   `json:"maintenanceMode"`
	MaintenanceMessage string `json:"maintenanceMessage"`
}
