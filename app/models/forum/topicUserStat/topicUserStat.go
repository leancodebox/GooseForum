package topicUserStat

import "time"

const tableName = "topic_user_stat"

type Entity struct {
	Id          uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	TopicId     uint64    `gorm:"column:topic_id;type:bigint unsigned;not null;uniqueIndex:idx_topic_user,priority:1;index:idx_active_topic_user,priority:1" json:"topicId"`
	UserId      uint64    `gorm:"column:user_id;type:bigint unsigned;not null;uniqueIndex:idx_topic_user,priority:2" json:"userId"`
	ReplyCount  uint32    `gorm:"column:reply_count;type:int unsigned;not null;default:0;index:idx_active_topic_user,priority:2" json:"replyCount"`
	LastReplyAt time.Time `gorm:"column:last_reply_at;type:datetime;not null;default:CURRENT_TIMESTAMP;index:idx_active_topic_user,priority:3" json:"lastReplyAt"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
