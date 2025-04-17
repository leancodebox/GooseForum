package userFollow

import (
	"time"
)

const tableName = "user_follow"

// pid
const pid = "id"

// fieldUserId
const fieldUserId = "user_id"

// fieldFollowUserId
const fieldFollowUserId = "follow_user_id"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id           uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                                                            //
	UserId       uint64    `gorm:"column:user_id;type:bigint unsigned;not null;uniqueIndex:uniq_user_follow,priority:1" json:"userId"`                                //
	FollowUserId uint64    `gorm:"column:follow_user_id;type:bigint unsigned;not null;uniqueIndex:uniq_user_follow,priority:2;index:idx_follow;" json:"followUserId"` //
	Status       int       `gorm:"column:status;type:int;not null;default:1;" json:"status"`                                                                          // 1 关注 0 取消关注
	CreatedAt    time.Time `gorm:"column:created_at;index;autoCreateTime;" json:"createdAt"`                                                                          //
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
