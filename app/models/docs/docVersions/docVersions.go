package docVersions

import (
	"gorm.io/gorm"
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

	CreatedAt time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`                 // 更新时间
	DeletedAt gorm.DeletedAt
}

// DirectoryItem 目录项结构
type DirectoryItem struct {
	Id          uint64          `json:"id"`
	Title       string          `json:"title"`
	Slug        string          `json:"slug"`
	Description string          `json:"description,omitempty"`
	Children    []DirectoryItem `json:"children,omitempty"`
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

func BuildSafeDescription(tree []DirectoryItem, contentList []DirectoryItem) []DirectoryItem {
	// 创建contentList的slug映射，用于快速查找
	contentMap := make(map[string]DirectoryItem)
	for _, content := range contentList {
		contentMap[content.Slug] = content
	}

	// 递归处理目录结构的函数
	var processDirectoryItems func(items []DirectoryItem) []DirectoryItem
	processDirectoryItems = func(items []DirectoryItem) []DirectoryItem {
		var result []DirectoryItem
		for _, dirItem := range items {
			if content, exists := contentMap[dirItem.Slug]; exists {
				// 使用contentList中的最新数据，但保留原有的Children结构
				updatedItem := content
				if len(dirItem.Children) > 0 {
					// 递归处理子项
					updatedItem.Children = processDirectoryItems(dirItem.Children)
				}
				result = append(result, updatedItem)
				// 从contentMap中删除已处理的项目
				delete(contentMap, dirItem.Slug)
			} else if len(dirItem.Children) > 0 {
				// 如果当前项不存在但有子项，递归处理子项
				processedChildren := processDirectoryItems(dirItem.Children)
				// 只有当处理后还有有效子项时才保留父项
				if len(processedChildren) > 0 {
					dirItem.Children = processedChildren
					result = append(result, dirItem)
				}
			}
			// 如果当前项不存在且没有子项，则跳过（不添加到result中）
		}
		return result
	}

	// 处理version.Directory，过滤掉不存在的内容
	validDirectoryItems := processDirectoryItems(tree)

	// 将contentList中剩余的（不在version.Directory中的）项目追加到最后
	for _, remainingContent := range contentMap {
		validDirectoryItems = append(validDirectoryItems, remainingContent)
	}

	// 转换为非指针类型的切片
	var resDirectory []DirectoryItem
	for _, item := range validDirectoryItems {
		resDirectory = append(resDirectory, item)
	}
	return resDirectory
}
