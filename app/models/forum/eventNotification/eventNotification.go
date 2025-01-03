package eventNotification

import (
	"time"
)

const tableName = "event_notification"

// pid
const pid = "id"

// fieldUserId
const fieldUserId = "user_id"

// fieldReceivedNotification
const fieldReceivedNotification = "received_notification"

// fieldEventType
const fieldEventType = "event_type"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id                   uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`              //
	UserId               string    `gorm:"column:user_id;type:varchar(255);" json:"userId"`                     //
	ReceivedNotification string    `gorm:"column:received_notification;type:text;" json:"receivedNotification"` //
	EventType            string    `gorm:"column:event_type;type:varchar(50);" json:"eventType"`                //
	CreatedAt            time.Time `gorm:"column:created_at;index;autoCreateTime;" json:"createdAt"`            //
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
