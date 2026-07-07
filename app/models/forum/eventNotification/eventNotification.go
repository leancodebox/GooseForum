package eventNotification

import (
	"time"
)

const tableName = "event_notification"

// Event Types
const (
	EventTypeComment   = "comment"    // 主题收到新 post
	EventTypePostReply = "post_reply" // post 回复通知
	EventTypeTopicPost = "topic_post" // 关注主题的新 post 通知
	EventTypeSystem    = "system"     // 系统通知
	EventTypeFollow    = "follow"     // 关注通知
	EventTypeBadge     = "badge"      // 徽章通知
)

const (
	TemplateComment   = "notifications.templates.comment"
	TemplatePostReply = "notifications.templates.postReply"
	TemplateTopicPost = "notifications.templates.topicPost"
	TemplateFollow    = "notifications.templates.follow"
	TemplateBadge     = "notifications.templates.badge"
)

// Future unread-scope design:
//
// Notifications are event records, but unread indicators in the UI often belong
// to a business scope instead of a single notification row. For example,
// comment/post_reply/topic_post notifications can all point to the same topic,
// and closing one of them should be able to clear the topic unread dot.
//
// When this is implemented, add a stable scope pair to Entity:
//   ScopeType string // topic, post, user, badge, system, ...
//   ScopeKey  string // topic id, post id, badge code, system batch key, ...
//
// Mark-as-read should stay permission anchored by notification id:
//   1. Client sends only notificationId.
//   2. Server loads id + current user id.
//   3. If the notification has ScopeType/ScopeKey, mark unread rows with the
//      same user_id + scope_type + scope_key as read.
//   4. If no scope is present, mark only the current notification as read.
//
// This avoids trusting client-provided scope values, keeps non-topic
// notifications independent, and lets different event types share one unread
// state when they point to the same business object.

// NotificationPayload 通知内容的基础结构
type NotificationPayload struct {
	Title          string                     `json:"title,omitempty"`
	Content        string                     `json:"content,omitempty"`
	TemplateKey    string                     `json:"templateKey,omitempty"`
	TemplateParams NotificationTemplateParams `json:"templateParams"`
	// 通用字段
	ActorId   uint64 `json:"actorId"`             // 触发者ID
	ActorName string `json:"actorName,omitempty"` // 触发者名称
	// Topic / post references
	TopicId    uint64 `json:"topicId,omitempty"`
	TopicTitle string `json:"topicTitle,omitempty"`
	PostId     uint64 `json:"postId,omitempty"`
	// 其他元数据
	Extra Extra `json:"metadata"`
}

type NotificationTemplateParams struct {
	Preview      string `json:"preview,omitempty"`
	FollowerName string `json:"followerName,omitempty"`
	BadgeCode    string `json:"badgeCode,omitempty"`
	BadgeName    string `json:"badgeName,omitempty"`
}

type Extra struct {
	FollowerName string `json:"followerName"`
	BadgeCode    string `json:"badgeCode,omitempty"`
	BadgeName    string `json:"badgeName,omitempty"`
	BadgeIconURL string `json:"badgeIconUrl,omitempty"`
	ProfileURL   string `json:"profileUrl,omitempty"`
}

type Entity struct {
	Id        uint64              `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_user_id_desc,priority:2;index:idx_user_read_id,priority:3" json:"id"`
	UserId    uint64              `gorm:"column:user_id;type:bigint;index:idx_user_id_event_type_read;index:idx_user_read;index:idx_user_id_desc,priority:1;index:idx_user_read_id,priority:1" json:"userId"` // 接收通知的用户ID
	Payload   NotificationPayload `gorm:"column:payload;type:json;serializer:json" json:"payload"`                                                                                                            // 通知内容(JSON)
	EventType string              `gorm:"column:event_type;type:varchar(16);index:idx_user_id_event_type_read;" json:"eventType"`                                                                             // 通知类型
	IsRead    bool                `gorm:"column:is_read;type:boolean;default:false;index:idx_user_id_event_type_read;index:idx_user_read;index:idx_user_read_id,priority:2" json:"isRead"`                    // 是否已读
	ReadAt    *time.Time          `gorm:"column:read_at;type:timestamp;null;" json:"readAt"`                                                                                                                  // 读取时间
	CreatedAt time.Time           `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`                                                                                                 // 创建时间
	UpdatedAt time.Time           `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`                                                                                                                 // 更新时间
}

func (itself *Entity) TableName() string {
	return tableName
}
