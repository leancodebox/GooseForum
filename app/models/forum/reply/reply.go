package reply

import (
	"time"
)

const tableName = "reply"

// pid
const pid = "id"

// fieldArticleId
const fieldArticleId = "article_id"

// fieldUserId
const fieldUserId = "user_id"

// fieldTargetId 目标id
const fieldTargetId = "target_id"

// fieldContent
const fieldContent = "content"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                      //
	ArticleId uint64    `gorm:"column:article_id;type:bigint unsigned;not null;default:0;" json:"articleId"` //
	UserId    uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`       //
	TargetId  uint64    `gorm:"column:target_id;type:bigint unsigned;not null;default:0;" json:"targetId"`   // 目标id
	Content   string    `gorm:"column:content;type:text;" json:"content"`                                    //
	CreatedAt time.Time `gorm:"column:created_at;index;autoCreateTime;" json:"createdAt"`                    //
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
