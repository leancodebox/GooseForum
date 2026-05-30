package articleWatch

import (
	"time"
)

const tableName = "article_watch"

const pid = "id"

const fieldUserId = "user_id"

const fieldArticleId = "article_id"

const fieldStatus = "status"

const fieldCreatedAt = "created_at"

const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	UserId    uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_article_watch,priority:1;index:idx_user_watch_list,priority:1;index:idx_article_watch_notify,priority:3" json:"userId"`
	ArticleId uint64    `gorm:"column:article_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_article_watch,priority:2;index:idx_article_watch_notify,priority:1" json:"articleId"`
	Status    int       `gorm:"column:status;type:int;not null;default:1;index:idx_user_watch_list,priority:2;index:idx_article_watch_notify,priority:2" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;index;index:idx_user_watch_list,priority:3" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
