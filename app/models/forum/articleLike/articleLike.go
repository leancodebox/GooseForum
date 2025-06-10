package articleLike

import (
	"time"
)

const tableName = "article_like"

// pid 主键
const pid = "id"

// fieldUserId
const fieldUserId = "user_id"

// fieldArticleId
const fieldArticleId = "article_id"

// fieldStatus 点赞状态（1:有效点赞 0:取消点赞）
const fieldStatus = "status"

// fieldUpdatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                                              // 主键
	UserId    uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_article,priority:1" json:"userId"`       // 用户
	ArticleId uint64    `gorm:"column:article_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_user_article,priority:2" json:"articleId"` //
	Status    int       `gorm:"column:status;type:int;not null;default:1;" json:"status"`                                                            // 点赞状态（1:有效点赞 0:取消点赞）
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`                                                        //
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;index;" json:"updatedAt"`
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
