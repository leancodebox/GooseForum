package friendshipLinks

import (
	"time"
)

const tableName = "friendship_links"

// pid 主键
const pid = "id"

// fieldSiteName 站点名
const fieldSiteName = "siteName"

// fieldSiteUrl 站点内容
const fieldSiteUrl = "siteUrl"

// fieldSiteLogo 站点logo
const fieldSiteLogo = "siteLogo"

// fieldSiteDesc 站点介绍
const fieldSiteDesc = "siteDesc"

// fieldContact 链接站长
const fieldContact = "contact"

// fieldWeight 权重
const fieldWeight = "weight"

// fieldStatus 状态 0 不展示 1 展示
const fieldStatus = "status"

// fieldLinkGroup 分类
const fieldLinkGroup = "link_group"

// fieldCreateTime
const fieldCreatedAt = "created_at"

// fieldUpdateTime
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                          // 主键
	SiteName  string    `gorm:"column:siteName;type:varchar(255);not null;default:0;" json:"siteName"`           // 站点名
	SiteUrl   string    `gorm:"column:siteUrl;type:varchar(255);not null;default:0;" json:"siteUrl"`             // 站点内容
	SiteLogo  string    `gorm:"column:siteLogo;type:varchar(255);not null;default:'';" json:"siteLogo"`          // 站点logo
	SiteDesc  string    `gorm:"column:siteDesc;type:varchar(255);not null;default:'';" json:"siteDesc"`          // 站点介绍
	Contact   string    `gorm:"column:contact;type:varchar(255);not null;default:'';" json:"contact"`            // 链接站长
	Weight    int       `gorm:"column:weight;type:int;not null;default:0;" json:"weight"`                        // 权重
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:0;" json:"status"`                    // 状态 0 不展示 1 展示
	LinkGroup string    `gorm:"column:link_group;type:varchar(32);not null;default:community;" json:"linkGroup"` // 分类
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;" json:"createdAt"`                              //
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;index;" json:"updatedAt"`
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
