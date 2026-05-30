package eventNotification

import (
	"time"
)

const tableName = "event_notification"

// Event Types
const (
	EventTypeComment        = "comment"         // 评论通知
	EventTypeReply          = "reply"           // 回复通知
	EventTypeArticleComment = "article_comment" // 关注文章评论通知
	EventTypeSystem         = "system"          // 系统通知
	EventTypeFollow         = "follow"          // 关注通知
	EventTypeBadge          = "badge"           // 徽章通知
)

const (
	TemplateComment        = "notifications.templates.comment"
	TemplateReply          = "notifications.templates.reply"
	TemplateArticleComment = "notifications.templates.articleComment"
	TemplateFollow         = "notifications.templates.follow"
	TemplateBadge          = "notifications.templates.badge"
)

// NotificationPayload 通知内容的基础结构
type NotificationPayload struct {
	Title          string                     `json:"title,omitempty"`          // 旧通知兼容
	Content        string                     `json:"content,omitempty"`        // 旧通知兼容
	TemplateKey    string                     `json:"templateKey,omitempty"`    // 前端 i18n 模板 key
	TemplateParams NotificationTemplateParams `json:"templateParams,omitempty"` // 前端 i18n 模板参数
	// 通用字段
	ActorId   uint64 `json:"actorId"`             // 触发者ID
	ActorName string `json:"actorName,omitempty"` // 触发者名称
	// 文章相关
	ArticleId    uint64 `json:"articleId,omitempty"`    // 相关文章ID
	ArticleTitle string `json:"articleTitle,omitempty"` // 文章标题
	// 评论相关
	CommentId uint64 `json:"commentId,omitempty"` // 评论ID
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
