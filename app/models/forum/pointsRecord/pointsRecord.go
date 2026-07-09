package pointsRecord

import (
	"time"
)

const tableName = "points_record"

// pid
const pid = "id"

// fieldUserId
const fieldUserId = "user_id"

// fieldAction
const fieldAction = "action"

// fieldPointsChange
const fieldPointsChange = "points_change"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

type Entity struct {
	Id           uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                   //
	UserId       uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`    //
	Action       string    `gorm:"column:action;type:varchar(64);not null;default:'';index;" json:"action"`  //
	PointsChange int64     `gorm:"column:points_change;type:bigint;not null;default:0;" json:"pointsChange"` //
	CreatedAt    time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`       //
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
