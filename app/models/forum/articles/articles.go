package articles

import (
	"time"
)

const tableName = "articles"

// pid
const pid = "id"

// fieldTitle
const fieldTitle = "title"

// fieldContent
const fieldContent = "content"

// fieldUserId
const fieldUserId = "user_id"

// fieldCreateTime
const fieldCreateTime = "create_time"

// fieldUpdateTime
const fieldUpdateTime = "update_time"

// fieldDeletedAt
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id         uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`           //
	Title      string     `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"` //
	Content    string     `gorm:"column:content;type:text;" json:"content"`                         //
	UserId     uint64     `gorm:"column:user_id;type:bigint;not null;default:0;" json:"userId"`     //
	CreateTime time.Time  `gorm:"column:create_time;type:datetime;not null;" json:"createTime"`     //
	UpdateTime time.Time  `gorm:"column:update_time;type:datetime;not null;" json:"updateTime"`     //
	DeletedAt  *time.Time `gorm:"column:deleted_at;type:datetime(3);" json:"deletedAt"`             //
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
