package articleUserAction

import "time"

const tableName = "article_user_action"

const fieldUserId = "user_id"
const fieldArticleId = "article_id"
const fieldLikedAt = "liked_at"
const fieldBookmarkedAt = "bookmarked_at"
const fieldWatchedAt = "watched_at"

type Entity struct {
	Id           uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_aus_user_id,priority:2" json:"id"`
	UserId       uint64     `gorm:"column:user_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_article_action,priority:1;index:idx_aus_user_id,priority:1;index:idx_aus_user_bookmark,priority:1;index:idx_aus_user_watch,priority:1;index:idx_aus_watch_notify,priority:2" json:"userId"`
	ArticleId    uint64     `gorm:"column:article_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_article_action,priority:2;index:idx_aus_watch_notify,priority:1" json:"articleId"`
	LikedAt      *time.Time `gorm:"column:liked_at" json:"likedAt"`
	BookmarkedAt *time.Time `gorm:"column:bookmarked_at;index:idx_aus_user_bookmark,priority:2" json:"bookmarkedAt"`
	WatchedAt    *time.Time `gorm:"column:watched_at;index:idx_aus_user_watch,priority:2;index:idx_aus_watch_notify,priority:3" json:"watchedAt"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime;<-:create" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
