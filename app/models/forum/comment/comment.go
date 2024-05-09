package comment

import (
	"time"
)

const tableName = "comment"

// pid
const pid = "id"

// fieldArticleId
const fieldArticleId = "article_id"

// fieldContent
const fieldContent = "content"

// fieldUserId
const fieldUserId = "user_id"

// fieldCreateTime
const fieldCreateTime = "create_time"

// fieldUpdateTime
const fieldUpdateTime = "update_time"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`             //
	ArticleId uint64    `gorm:"column:article_id;type:bigint;not null;default:0;" json:"articleId"` //
	Content   string    `gorm:"column:content;type:text;" json:"content"`                           //
	UserId    uint64    `gorm:"column:user_id;type:bigint;not null;default:0;" json:"userId"`       //
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;" json:"createdAt"`                  //
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;" json:"updatedAt"`                  //
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
