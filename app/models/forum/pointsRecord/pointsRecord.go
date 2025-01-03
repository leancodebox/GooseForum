package pointsRecord

import (
	"time"
)

const tableName = "points_record"

// pid
const pid = "id"

// fieldUserId
const fieldUserId = "user_id"

// fieldChangeReason
const fieldChangeReason = "change_reason"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

type Entity struct {
	Id           uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                //
	UserId       uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"` //
	ChangeReason string    `gorm:"column:change_reason;type:varchar(255);" json:"changeReason"`           //
	CreatedAt    time.Time `gorm:"column:created_at;index;autoCreateTime;" json:"createdAt"`              //
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
