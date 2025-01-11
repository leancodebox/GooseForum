package articleCategoryRs

import (
	"time"
)

const tableName = "article_category_rs"

// pid
const pid = "id"

// fieldArticleId
const fieldArticleId = "article_id"

// fieldArticleCategoryId
const fieldArticleCategoryId = "article_category_id"

// fieldEffective
const fieldEffective = "effective"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id                uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                 //
	ArticleId         uint64    `gorm:"column:article_id;type:bigint unsigned;not null;index:idx_article_id;" json:"articleId"` //
	ArticleCategoryId uint64    `gorm:"column:article_category_id;type:bigint unsigned;not null;" json:"articleCategoryId"`     //
	Effective         int       `gorm:"column:effective;type:int;not null;default:0;" json:"effective"`                         //
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime;" json:"createdAt"`                                     //
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
