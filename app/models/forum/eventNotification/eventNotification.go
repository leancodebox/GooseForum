package eventNotification

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

const tableName = "event_notification"

// Event Types
const (
	EventTypeComment  = "comment"   // 评论通知
	EventTypeReply    = "reply"     // 回复通知
	EventTypeSystem   = "system"    // 系统通知
	EventTypeFollow   = "follow"    // 关注通知
)

// NotificationPayload 通知内容的基础结构
type NotificationPayload struct {
	Title   string                 `json:"title"`              // 通知标题
	Content string                 `json:"content"`            // 通知内容
	Extra   map[string]interface{} `json:"extra,omitempty"`   // 额外数据
}

// Value 实现 driver.Valuer 接口
func (p NotificationPayload) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口
func (p *NotificationPayload) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), p)
}

type Entity struct {
	Id        uint64             `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	UserId    uint64             `gorm:"column:user_id;type:bigint;index;" json:"userId"`                    // 接收通知的用户ID
	Payload   NotificationPayload `gorm:"column:payload;type:json;" json:"payload"`                          // 通知内容(JSON)
	EventType string             `gorm:"column:event_type;type:varchar(50);index;" json:"eventType"`         // 通知类型
	IsRead    bool               `gorm:"column:is_read;type:boolean;default:false;index;" json:"isRead"`     // 是否已读
	ReadAt    *time.Time         `gorm:"column:read_at;type:timestamp;null;" json:"readAt"`                 // 读取时间
	CreatedAt time.Time          `gorm:"column:created_at;index;autoCreateTime;" json:"createdAt"`          // 创建时间
	UpdatedAt time.Time          `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`                // 更新时间
}

func (itself *Entity) TableName() string {
	return tableName
}
