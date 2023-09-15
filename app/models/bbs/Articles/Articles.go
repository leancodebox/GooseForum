package Articles

import (
	"time"

	"gorm.io/gorm"
)

const tableName = "articles"
const pid = "id"
const fieldContent = "content"
const fieldUserId = "user_id"
const fieldCreateTime = "create_time"
const fieldUpdateTime = "update_time"

type Entity struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"` // 主键
	Title      string    `gorm:"type:varchar(512);not null;default:'';" json:"title"`
	Content    string    `gorm:"column:content;type:text;default:'';" json:"content"`                              //
	UserId     uint64    `gorm:"column:user_id;type:bigint;not null;default:0;" json:"userId"`                     //
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime;type:datetime;not null;" json:"createTime"`      //
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime:true;type:datetime;not null;" json:"updateTime"` //
	DeletedAt  gorm.DeletedAt
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
