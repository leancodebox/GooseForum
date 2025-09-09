package userUnreadCounts

import (
	"time"
)

const tableName = "user_unread_counts"

// pid id
const pid = "user_id"

// fieldReplyCount 回复消息
const fieldReplyCount = "reply_count"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

type Entity struct {
	UserId     uint64    `gorm:"primaryKey;column:user_id;autoIncrement;not null;" json:"userId"`    // id
	ReplyCount int       `gorm:"column:reply_count;type:int;not null;default:0;" json:"replyCount"`  // 回复消息
	LikeCount  int       `gorm:"column:like_count;type:int;not null;default:0;" json:"likeCount"`    // 点赞消息
	CreatedAt  time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` //
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
