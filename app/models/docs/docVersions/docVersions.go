package docVersions

import (
	"time"
)

const tableName = "doc_versions"

// pid 版本ID
const pid = "id"

// fieldName 版本名称
const fieldName = "name"

// fieldSlug 版本标识符
const fieldSlug = "slug"

// fieldDescription 版本描述
const fieldDescription = "description"

// fieldIsDefault 是否默认版本(0:否 1:是)
const fieldIsDefault = "is_default"

// fieldIsPublished 是否发布(0:草稿 1:已发布)
const fieldIsPublished = "is_published"

// fieldSortOrder 排序权重
const fieldSortOrder = "sort_order"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

// fieldProjectId 项目ID
const fieldProjectId = "project_id"

// fieldDeletedAt 删除时间
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id          uint64          `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                  // 版本ID
	ProjectId   uint64          `gorm:"column:project_id;type:bigint;not null;" json:"projectId"`                // 项目ID
	Name        string          `gorm:"column:name;type:varchar(50);not null;default:'';" json:"name"`           // 版本名称
	Slug        string          `gorm:"column:slug;type:varchar(50);not null;default:'';" json:"slug"`           // 版本标识符
	Description string          `gorm:"column:description;type:text;" json:"description"`                        // 版本描述
	Status      int8            `gorm:"column:status;type:tinyint;not null;default:1;" json:"status"`            // 状态(1:活跃 2:维护 3:废弃)
	IsDefault   int8            `gorm:"column:is_default;type:tinyint;not null;default:0;" json:"isDefault"`     // 是否默认版本(0:否 1:是)
	IsPublished int8            `gorm:"column:is_published;type:tinyint;not null;default:0;" json:"isPublished"` // 是否发布(0:草稿 1:已发布)
	SortOrder   int             `gorm:"column:sort_order;type:int;not null;default:0;" json:"sortOrder"`         // 排序权重
	Directory   []DirectoryItem `gorm:"column:directory;type:text;serializer:json" json:"directory"`

	CreatedAt time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`                 // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"`                  // 删除时间
}

// DirectoryItem 目录项结构
type DirectoryItem struct {
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Description string           `json:"description,omitempty"`
	Children    []*DirectoryItem `json:"children,omitempty"`
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
