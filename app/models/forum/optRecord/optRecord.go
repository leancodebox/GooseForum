package optRecord

import (
	"time"
)

const tableName = "opt_record"

// pid
const pid = "id"

// fieldOptUserId
const fieldOptUserId = "opt_user_id"

// fieldOptType
const fieldOptType = "opt_type"

// fieldTargetType
const fieldTargetType = "target_type"

// fieldTargetId
const fieldTargetId = "target_id"

// fieldOptInfo
const fieldOptInfo = "opt_info"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

type Entity struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                       //
	OptUserId  uint64    `gorm:"column:opt_user_id;type:bigint unsigned;not null;default:0;" json:"optUserId"` //
	OptType    int       `gorm:"column:opt_type;type:int;not null;default:0;" json:"optType"`                  //
	TargetType int       `gorm:"column:target_type;type:int;not null;default:0;" json:"targetType"`            //
	TargetId   string    `gorm:"column:target_id;type:varchar(255);not null;default:'';" json:"targetId"`      //
	OptInfo    string    `gorm:"column:opt_info;type:text;" json:"optInfo"`                                    //
	CreatedAt  time.Time `gorm:"column:created_at;index;autoCreateTime;" json:"createdAt"`                     //
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
