package docProjects

import (
	"gorm.io/gorm"
	"time"
)

const tableName = "doc_projects"

// pid 项目ID
const pid = "id"

// fieldName 项目名称
const fieldName = "name"

// fieldSlug 项目标识符
const fieldSlug = "slug"

// fieldDescription 项目描述
const fieldDescription = "description"

// fieldRepositoryUrl 仓库地址
const fieldRepositoryUrl = "repository_url"

// fieldHomepageUrl 项目主页
const fieldHomepageUrl = "homepage_url"

// fieldLogoUrl 项目Logo
const fieldLogoUrl = "logo_url"

// fieldStatus 状态(1:活跃 2:维护 3:废弃)
const fieldStatus = "status"

// fieldIsPublic 是否公开(0:私有 1:公开)
const fieldIsPublic = "is_public"

// fieldOwnerId 项目所有者ID
const fieldOwnerId = "owner_id"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt 删除时间
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id          uint64 `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                  // 项目ID
	Name        string `gorm:"column:name;type:varchar(100);not null;default:'';" json:"name"`          // 项目名称
	Slug        string `gorm:"column:slug;type:varchar(100);not null;default:'';" json:"slug"`          // 项目标识符
	Description string `gorm:"column:description;type:text;" json:"description"`                        // 项目描述
	LogoUrl     string `gorm:"column:logo_url;type:varchar(255);not null;default:'';" json:"logoUrl"`   // 项目Logo
	Status      int8   `gorm:"column:status;type:tinyint;not null;default:1;" json:"status"`            // 状态(1:活跃 2:维护 3:废弃)
	IsPublic    int8   `gorm:"column:is_public;type:tinyint;not null;default:1;" json:"isPublic"`       // 是否公开(0:私有 1:公开)
	OwnerId     uint64 `gorm:"column:owner_id;type:bigint unsigned;not null;default:0;" json:"ownerId"` // 项目所有者ID

	CreatedAt time.Time      `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`                 // 更新时间
	DeletedAt gorm.DeletedAt // 删除时间
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
