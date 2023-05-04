package Comment

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

type Comment struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                                     // 主键
	ArticleId  uint64    `gorm:"column:article_id;type:bigint;not null;default:0;" json:"articleId"`                                         //
	Content    string    `gorm:"column:content;type:text;default:'';" json:"content"`                                                        //
	UserId     uint64    `gorm:"column:user_id;type:bigint;not null;default:0;" json:"userId"`                                               //
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP;" json:"createTime"`                     //
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime:true;type:datetime;not null;default:CURRENT_TIMESTAMP;" json:"updateTime"` //
}

// func (itself *Comment) BeforeSave(tx *gorm.DB) (err error) {}
// func (itself *Comment) BeforeCreate(tx *gorm.DB) (err error) {}
// func (itself *Comment) AfterCreate(tx *gorm.DB) (err error) {}
// func (itself *Comment) BeforeUpdate(tx *gorm.DB) (err error) {}
// func (itself *Comment) AfterUpdate(tx *gorm.DB) (err error) {}
// func (itself *Comment) AfterSave(tx *gorm.DB) (err error) {}
// func (itself *Comment) BeforeDelete(tx *gorm.DB) (err error) {}
// func (itself *Comment) AfterDelete(tx *gorm.DB) (err error) {}
// func (itself *Comment) AfterFind(tx *gorm.DB) (err error) {}

func (Comment) TableName() string {
	return tableName
}
