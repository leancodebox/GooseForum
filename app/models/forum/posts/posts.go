package posts

import (
	"time"

	"gorm.io/gorm"
)

const tableName = "posts"

type Entity struct {
	PublishedAt       *time.Time     `gorm:"column:published_at" json:"-"`
	Id                uint64         `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_posts_topic_id,priority:2;;index:idx_posts_review,priority:3" json:"id"`
	TopicId           uint64         `gorm:"column:topic_id;not null;default:0;index:idx_posts_topic_created,priority:1;uniqueIndex:idx_posts_topic_no,priority:1;index:idx_posts_topic_id,priority:1;index:idx_posts_topic_process,priority:1;" json:"topicId"`
	PostNo            uint64         `gorm:"column:post_no;not null;default:0;uniqueIndex:idx_posts_topic_no,priority:2;" json:"postNo"`
	UserId            uint64         `gorm:"column:user_id;not null;default:0;index;" json:"userId"`
	ReplyToPostId     uint64         `gorm:"column:reply_to_post_id;not null;default:0;" json:"replyToPostId"`
	Content           string         `gorm:"column:content;type:text;" json:"content"`
	RenderedHTML      string         `gorm:"column:rendered_html;type:text;" json:"renderedHTML"`
	RenderedVersion   uint32         `gorm:"column:rendered_version;not null;default:0;" json:"renderedVersion"`
	ProcessStatus     int8           `gorm:"column:process_status;not null;default:0;index:idx_posts_topic_process,priority:2;" json:"processStatus"`
	ModerationStatus  string         `gorm:"column:moderation_status;type:varchar(16);not null;default:'none';index:idx_posts_review,priority:1" json:"moderationStatus"`
	ModerationVersion uint64         `gorm:"column:moderation_version;not null;default:0" json:"-"`
	ModerationReason  string         `gorm:"column:moderation_reason;type:varchar(512);not null;default:''" json:"-"`
	ModeratedAt       *time.Time     `gorm:"column:moderated_at" json:"-"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime;<-:create;index:idx_posts_topic_created,priority:2;" json:"createdAt"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime;;index:idx_posts_review,priority:2" json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `json:"-"`
}

func (itself *Entity) TableName() string {
	return tableName
}

// Older replies predate moderation and were already published. Preserve that
// fact when they are first edited under the new schema.
func (e Entity) WasPublished() bool {
	return e.PublishedAt != nil || (e.Id != 0 && (e.ModerationVersion == 0 || e.ModerationStatus == "approved"))
}
