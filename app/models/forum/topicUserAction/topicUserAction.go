package topicUserAction

import "time"

const tableName = "topic_user_action"

type Entity struct {
	Id           uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_tua_user_id,priority:2" json:"id"`
	UserId       uint64     `gorm:"column:user_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_topic_action,priority:1;index:idx_tua_user_id,priority:1;index:idx_tua_user_bookmark,priority:1;index:idx_tua_user_watch,priority:1;index:idx_tua_watch_notify,priority:2" json:"userId"`
	TopicId      uint64     `gorm:"column:topic_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_topic_action,priority:2;index:idx_tua_watch_notify,priority:1" json:"topicId"`
	LikedAt      *time.Time `gorm:"column:liked_at" json:"likedAt"`
	BookmarkedAt *time.Time `gorm:"column:bookmarked_at;index:idx_tua_user_bookmark,priority:2" json:"bookmarkedAt"`
	WatchedAt    *time.Time `gorm:"column:watched_at;index:idx_tua_user_watch,priority:2;index:idx_tua_watch_notify,priority:3" json:"watchedAt"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime;<-:create" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
