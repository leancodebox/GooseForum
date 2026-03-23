package imConversations

import (
	"time"
)

const tableName = "im_conversations"

// pid 主键
const pid = "id"

// fieldType 1 单聊 (C2C)
const fieldType = "type"

// fieldLastMsgContent 最后一条消息预览
const fieldLastMsgContent = "last_msg_content"

// fieldLastMsgTime 用于排序
const fieldLastMsgTime = "last_msg_time"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

type Entity struct {
	Id             uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                    // 主键
	Type           int       `gorm:"column:type;type:tinyint(1);not null;default:1;" json:"type"`                               // 1 单聊 (C2C)
	LastMsgContent string    `gorm:"column:last_msg_content;type:varchar(255);default:'';" json:"lastMsgContent"`               // 最后一条消息预览
	LastMsgTime    time.Time `gorm:"column:last_msg_time;type:datetime;not null;default:CURRENT_TIMESTAMP;" json:"lastMsgTime"` // 用于排序
	CreatedAt      time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`                        //
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
