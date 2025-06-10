package articleCollection

import (
	"time"
)

const tableName = "article_collection"

// pid 主键
const pid = "id"

// fieldUserId 用户ID
const fieldUserId = "user_id"

// fieldArticleId 文章ID
const fieldArticleId = "article_id"

// fieldStatus 有效收藏 1 无效收藏 0
const fieldStatus = "status"

// fieldCreateTime 收藏时间
const fieldCreateTime = "create_time"

// fieldUpdateTime 更新时间
const fieldUpdateTime = "update_time"

type Entity struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                                             // 主键
	UserId     uint64    `gorm:"column:user_id;type:bigint unsigned;not null;" json:"userId"`                                                        // 用户ID
	ArticleId  uint64    `gorm:"column:article_id;type:bigint unsigned;not null;" json:"articleId"`                                                  // 文章ID
	Status     int8      `gorm:"column:status;type:tinyint;not null;default:0;" json:"status"`                                                       // 有效收藏 1 无效收藏 0
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP;" json:"createTime"`                             // 收藏时间
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;" json:"updateTime"` // 更新时间
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
