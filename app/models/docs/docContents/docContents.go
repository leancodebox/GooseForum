package docContents

import (
	"time"
)

const tableName = "doc_contents"

// pid 内容ID
const pid = "id"

// fieldTitle 文档标题
const fieldTitle = "title"

// fieldSlug 文档标识符
const fieldSlug = "slug"

// fieldContent 文档内容(Markdown)
const fieldContent = "content"

// fieldContentHtml 渲染后的HTML内容
const fieldContentHtml = "content_html"

// fieldExcerpt 文档摘要
const fieldExcerpt = "excerpt"

// fieldToc 目录结构(JSON)
const fieldToc = "toc"

// fieldMetaKeywords SEO关键词
const fieldMetaKeywords = "meta_keywords"

// fieldMetaDescription SEO描述
const fieldMetaDescription = "meta_description"

// fieldIsPublished 是否发布(0:草稿 1:已发布)
const fieldIsPublished = "is_published"

// fieldSortOrder 排序权重
const fieldSortOrder = "sort_order"

// fieldAuthorId 作者ID
const fieldAuthorId = "author_id"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt 删除时间
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id              uint64 `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                    // 内容ID
	Title           string `gorm:"column:title;type:varchar(200);not null;default:'';" json:"title"`          // 文档标题
	Slug            string `gorm:"column:slug;type:varchar(200);not null;default:'';" json:"slug"`            // 文档标识符
	Content         string `gorm:"column:content;type:longtext;" json:"content"`                              // 文档内容(Markdown)
	ContentHtml     string `gorm:"column:content_html;type:longtext;" json:"contentHtml"`                     // 渲染后的HTML内容
	Excerpt         string `gorm:"column:excerpt;type:text;" json:"excerpt"`                                  // 文档摘要
	Toc             string `gorm:"column:toc;type:json;" json:"toc"`                                          // 目录结构(JSON)
	MetaKeywords    string `gorm:"column:meta_keywords;type:varchar(255);" json:"metaKeywords"`               // SEO关键词
	MetaDescription string `gorm:"column:meta_description;type:text;" json:"metaDescription"`                 // SEO描述
	IsPublished     int8   `gorm:"column:is_published;type:tinyint;not null;default:0;" json:"isPublished"`   // 是否发布(0:草稿 1:已发布)
	SortOrder       int    `gorm:"column:sort_order;type:int;not null;default:0;" json:"sortOrder"`           // 排序权重
	AuthorId        uint64 `gorm:"column:author_id;type:bigint unsigned;not null;default:0;" json:"authorId"` // 作者ID

	CreatedAt time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`                 // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"`                  // 删除时间
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
