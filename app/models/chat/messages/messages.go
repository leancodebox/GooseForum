package messages

import (
	"time"
)

const tableName = "messages"

// pid 主键
const pid = "id"

// fieldConvId 所属对话ID
const fieldConvId = "conv_id"

// fieldSenderId 发送者ID
const fieldSenderId = "sender_id"

// fieldContent 消息文本
const fieldContent = "content"

// fieldMsgType 消息类型: 1文本, 2图片, 3语音等
const fieldMsgType = "msg_type"

// fieldIsRead 该消息是否被对方读取
const fieldIsRead = "is_read"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                    // 主键
	ConvId    uint64    `gorm:"column:conv_id;type:bigint unsigned;not null;default:0;index:idx_conv_time,priority:1" json:"convId"` // 所属对话ID
	SenderId  uint64    `gorm:"column:sender_id;type:bigint unsigned;not null;default:0;" json:"senderId"`                 // 发送者ID
	Content   string    `gorm:"column:content;type:text;not null;" json:"content"`                                         // 消息文本
	MsgType   int8      `gorm:"column:msg_type;type:tinyint;not null;default:1;" json:"msgType"`                           // 消息类型: 1文本, 2图片, 3语音等
	IsRead    int       `gorm:"column:is_read;type:tinyint(1);not null;default:0;" json:"isRead"`                          // 该消息是否被对方读取
	CreatedAt time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;index:idx_conv_time,priority:2" json:"createdAt"`    //
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
