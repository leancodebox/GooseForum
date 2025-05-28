package applySheet

import (
	"time"
)

const tableName = "apply_sheet"

// pid 主键
const pid = "id"

// fieldUserId
const fieldUserId = "user_id"

// fieldType
const fieldType = "type"

// fieldStatus 状态
const fieldStatus = "status"

// fieldTitle 标题
const fieldTitle = "title"

// fieldContent 具体内容
const fieldContent = "content"

// fieldCreateTime
const fieldCreatedAt = "created_at"

// fieldUpdateTime
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id            uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                // 主键
	UserId        uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"` //
	ApplyUserInfo string    `gorm:"column:apply_user_info;type:text;" json:"applyUserInfo"`                //
	Type          SheetType `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`              //
	Status        int8      `gorm:"column:status;type:tinyint;not null;default:1;" json:"status"`          // 状态
	Title         string    `gorm:"column:title;type:varchar(255);not null;default:'';" json:"title"`      // 标题
	Content       string    `gorm:"column:content;type:text;" json:"content"`                              // 具体内容
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`          //
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

type SheetType int8

const (
	ApplyAddLink SheetType = iota + 1
)

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
