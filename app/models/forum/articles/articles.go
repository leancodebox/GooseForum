package articles

import (
	"time"
)

const tableName = "articles"

// pid
const pid = "id"

// fieldTitle
const fieldTitle = "title"

// fieldContent
const fieldContent = "content"

// fieldType 文章类型：0 博文，1教程，2问答，3分享
const fieldType = "type"

// fieldUserId
const fieldUserId = "user_id"

// fieldArticleStatus 文章状态：0 草稿 1 发布
const fieldArticleStatus = "article_status"

// fieldProcessStatus 管理状态：0 正常 1 封禁
const fieldProcessStatus = "process_status"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id            uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                      //
	Title         string     `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"`            //
	Content       string     `gorm:"column:content;type:text;" json:"content"`                                    //
	Type          int8       `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`                    // 文章类型：0 博文，1教程，2问答，3分享
	UserId        uint64     `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`       //
	ArticleStatus int8       `gorm:"column:article_status;type:tinyint;not null;default:0;" json:"articleStatus"` // 文章状态：0 草稿 1 发布
	ProcessStatus int8       `gorm:"column:process_status;type:tinyint;not null;default:0;" json:"processStatus"` // 管理状态：0 正常 1 封禁
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime;" json:"createdAt"`                          //
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"` //
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
