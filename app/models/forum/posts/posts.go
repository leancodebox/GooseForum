package posts

import (
	"time"

	"gorm.io/gorm"
)

const tableName = "posts"

type Entity struct {
	Id              uint64         `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_posts_topic_id,priority:2;" json:"id"`
	TopicId         uint64         `gorm:"column:topic_id;type:bigint unsigned;not null;default:0;index:idx_posts_topic_created,priority:1;uniqueIndex:idx_posts_topic_no,priority:1;index:idx_posts_topic_id,priority:1;index:idx_posts_topic_process,priority:1;" json:"topicId"`
	PostNo          uint64         `gorm:"column:post_no;type:bigint unsigned;not null;default:0;uniqueIndex:idx_posts_topic_no,priority:2;" json:"postNo"`
	UserId          uint64         `gorm:"column:user_id;type:bigint unsigned;not null;default:0;index;" json:"userId"`
	ReplyToPostId   uint64         `gorm:"column:reply_to_post_id;type:bigint unsigned;not null;default:0;" json:"replyToPostId"`
	Content         string         `gorm:"column:content;type:text;" json:"content"`
	RenderedHTML    string         `gorm:"column:rendered_html;type:text;" json:"renderedHTML"`
	RenderedVersion uint32         `gorm:"column:rendered_version;type:bigint unsigned;not null;default:0;" json:"renderedVersion"`
	ProcessStatus   int8           `gorm:"column:process_status;type:tinyint;not null;default:0;index:idx_posts_topic_process,priority:2;" json:"processStatus"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime;<-:create;index:idx_posts_topic_created,priority:2;" json:"createdAt"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `json:"-"`
}

func (itself *Entity) TableName() string {
	return tableName
}
