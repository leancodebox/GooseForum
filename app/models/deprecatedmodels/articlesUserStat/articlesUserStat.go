package articlesUserStat

import (
	"time"
)

const tableName = "articles_user_stat"

// pid id
const pid = "id"

const filedArticleId = "article_id"

// fieldUserId 用户id
const fieldUserId = "user_id"

// fieldReplayCount 回复量
const fieldReplyCount = "reply_count"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id uint64 `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	// 联合唯一索引 idx_article_user 用于 Upsert (On Conflict) 定位
	// 联合索引 idx_active_query 用于高性能获取前 5 名活跃用户
	ArticleId uint64 `gorm:"column:article_id;type:bigint unsigned;not null;uniqueIndex:idx_article_user,priority:1;index:idx_active_query,priority:1" json:"articleId"`
	UserId    uint64 `gorm:"column:user_id;type:bigint unsigned;not null;uniqueIndex:idx_article_user,priority:2" json:"userId"`

	// 计数器和时间，用于排序
	ReplyCount  uint32    `gorm:"column:reply_count;type:int unsigned;not null;default:0;index:idx_active_query,priority:2" json:"replyCount"`
	LastReplyAt time.Time `gorm:"column:last_reply_at;type:datetime;not null;default:CURRENT_TIMESTAMP;index:idx_active_query,priority:3" json:"lastReplyAt"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
