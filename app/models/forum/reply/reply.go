package reply

import (
	"time"

	"gorm.io/gorm"
)

const tableName = "reply"

// pid
const pid = "id"

// fieldArticleId
const fieldArticleId = "article_id"

// fieldUserId
const fieldUserId = "user_id"

// fieldReplyNo
const fieldReplyNo = "reply_no"

// fieldTargetId 目标id
const fieldTargetId = "target_id"

// fieldContent
const fieldContent = "content"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id              uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_reply_article_id,priority:2;" json:"id"`                                                                                            //
	ArticleId       uint64    `gorm:"column:article_id;type:bigint unsigned;not null;default:0;index:idx_reply_article_created;index:idx_reply_article_id,priority:1;index:idx_reply_article_no,priority:1;" json:"articleId"` //
	ReplyNo         uint64    `gorm:"column:reply_no;type:bigint unsigned;not null;default:0;index:idx_reply_article_no,priority:2;" json:"replyNo"`                                                                           // 文章内稳定回复序号
	UserId          uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;index;" json:"userId"`                                                                                                             //
	TargetId        uint64    `gorm:"column:target_id;type:bigint unsigned;not null;default:0;" json:"targetId"`                                                                                                               // 目标id
	Content         string    `gorm:"column:content;type:text;" json:"content"`                                                                                                                                                //
	RenderedHTML    string    `gorm:"column:rendered_html;type:text;" json:"renderedHTML"`                                                                                                                                     // md 渲染后数据
	RenderedVersion uint32    `gorm:"column:rendered_version;type:bigint unsigned;not null;default:0;" json:"renderedVersion"`                                                                                                 // md 的渲染器版本
	ReplyId         uint64    `gorm:"column:reply_id;type:bigint;not null;default:0;" json:"replyId"`                                                                                                                          //
	CreatedAt       time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;index:idx_reply_article_created;" json:"createdAt"`                                                                                      //
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	gorm.DeletedAt
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
