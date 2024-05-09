package comment

import (
	"time"
)

const tableName = "comment"
const pid = "id"
const fieldArticleId = "article_id"
const fieldContent = "content"
const fieldUserId = "user_id"
const fieldCreateTime = "create_time"
const fieldUpdateTime = "update_time"

type Entity struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                                     // 主键
	ArticleId  uint64    `gorm:"column:article_id;type:bigint;not null;default:0;" json:"articleId"`                                         //
	Content    string    `gorm:"column:content;type:text;" json:"content"`                                                                   //
	UserId     uint64    `gorm:"column:user_id;type:bigint;not null;default:0;" json:"userId"`                                               //
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP;" json:"createTime"`                     //
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime:true;type:datetime;not null;default:CURRENT_TIMESTAMP;" json:"updateTime"` //
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

func (Entity) TableName() string {
	return tableName
}
