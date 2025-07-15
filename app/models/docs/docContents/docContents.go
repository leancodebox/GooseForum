package docContents

import (
	"gorm.io/gorm"
	"time"
)

const tableName = "doc_contents"

// pid 内容ID
const pid = "id"

// fieldTitle 文档标题
const fieldTitle = "title"

// fieldVersionId 版本id
const fieldVersionId = "version_id"

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
	Id                 uint64 `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                         // 内容ID
	VersionId          uint64 `gorm:"column:version_id;not null;default:0;uniqueIndex:version_id,slug:1" json:"versionId"`            // 版本
	Title              string `gorm:"column:title;type:varchar(200);not null;default:'';" json:"title"`                               // 文档标题
	Slug               string `gorm:"column:slug;type:varchar(200);not null;default:'';uniqueIndex:version_id,slug:2" json:"slug"`    // 文档标识符
	OriginContent      string `gorm:"column:origin_content;type:text;" json:"OriginContent"`                                          // 文档内容(Markdown)
	Content            string `gorm:"column:content;type:text;" json:"content"`                                                       // 文档内容(Markdown)
	ContentHtml        string `gorm:"column:content_html;type:text;" json:"contentHtml"`                                              // 渲染后的HTML内容
	ContentHtmlVersion uint32 `gorm:"column:content_html_version;type:bigint unsigned;not null;default:0;" json:"contentHtmlVersion"` //md 的渲染器版本
	Toc                string `gorm:"column:toc;type:json;" json:"toc"`                                                               // 目录结构(JSON)
	IsPublished        int8   `gorm:"column:is_published;type:tinyint;not null;default:0;" json:"isPublished"`                        // 是否发布(0:草稿 1:已发布)
	SortOrder          int    `gorm:"column:sort_order;type:int;not null;default:0;" json:"sortOrder"`                                // 排序权重

	LikeCount uint64 `gorm:"column:like_count;type:bigint unsigned;not null;default:0;" json:"likeCount"`       // 喜欢数量
	ViewCount uint64 `gorm:"column:view_count;index;type:bigint unsigned;not null;default:0;" json:"viewCount"` // 访问数量

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
