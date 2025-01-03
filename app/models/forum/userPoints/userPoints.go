package userPoints

import (
	"time"
)

const tableName = "user_points"

// pid
const pid = "user_id"

// fieldCurrentPoints
const fieldCurrentPoints = "current_points"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

type Entity struct {
	UserId        uint64    `gorm:"primaryKey;column:user_id;autoIncrement;not null;" json:"userId"`            //
	CurrentPoints int64     `gorm:"column:current_points;type:bigint;not null;default:0;" json:"currentPoints"` //
	CreatedAt     time.Time `gorm:"column:created_at;index;autoCreateTime;" json:"createdAt"`                   //
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
