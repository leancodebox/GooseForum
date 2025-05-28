package filedata

import (
	"time"
)

const tableName = "file_data"
const pid = "id"
const fieldName = "name"
const fieldType = "type"
const fieldCreateTime = "create_time"
const fieldUpdateTime = "update_time"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                      // 主键
	Name      string    `gorm:"column:name;uniqueIndex;type:varchar(256);not null;" json:"name"`             // 添加唯一索引
	Type      string    `gorm:"column:assert_type;index;type:varchar(64);not null;" json:"type"`             //
	Data      []byte    `gorm:"column:content;type:BLOB;" json:"data"`                                       // 内容
	UserId    uint64    `gorm:"column:user_id;index;type:bigint unsigned;not null;default:0;" json:"userId"` //
	CreatedAt time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`          //
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
