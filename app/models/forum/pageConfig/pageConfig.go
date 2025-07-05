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
)

var PageTypeList = []string{
	FriendShipLinks,
	WebSettings,
}

type WebSettingsConfig struct {
	ExternalLinks string `json:"externalLinks,omitempty"`
	CustomJS      string `json:"customJS,omitempty"`
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
