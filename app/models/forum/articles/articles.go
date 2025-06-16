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

// fieldReplyCount 回复量
const fieldReplyCount = "view_count"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id              uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                  //
	Title           string     `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"`                        //
	Content         string     `gorm:"column:content;type:text;" json:"content"`                                                //
	Description     string     `gorm:"column:description;type:varchar(255);not null;default:'';" json:"description"`            // 文章描述，用于SEO
	RenderedHTML    string     `gorm:"column:rendered_html;type:text;" json:"renderedHTML"`                                     //md 渲染后数据
	RenderedVersion uint32     `gorm:"column:rendered_version;type:bigint unsigned;not null;default:0;" json:"renderedVersion"` //md 的渲染器版本
	Type            int8       `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`                                // 文章类型：0 博文，1教程，2问答，3分享
	UserId          uint64     `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`                   //
	ArticleStatus   int8       `gorm:"column:article_status;type:tinyint;not null;default:0;" json:"articleStatus"`             // 文章状态：0 草稿 1 发布
	ProcessStatus   int8       `gorm:"column:process_status;type:tinyint;not null;default:0;" json:"processStatus"`             // 管理状态：0 正常 1 封禁
	LikeCount       uint64     `gorm:"column:like_count;type:bigint unsigned;not null;default:0;" json:"likeCount"`             // 访问数量
	ViewCount       uint64     `gorm:"column:view_count;index;type:bigint unsigned;not null;default:0;" json:"viewCount"`       // 访问数量
	ReplyCount      uint64     `gorm:"column:reply_count;type:bigint unsigned;not null;default:0;" json:"replyCount"`           // 访问数量
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`                            //
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime;index;" json:"updatedAt"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"` //
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

type SmallEntity struct {
	Id            uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                        //
	Title         string     `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"`              //
	Type          int8       `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`                      // 文章类型：0 博文，1教程，2问答，3分享
	UserId        uint64     `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`         //
	ArticleStatus int8       `gorm:"column:article_status;type:tinyint;not null;default:0;" json:"articleStatus"`   // 文章状态：0 草稿 1 发布
	ProcessStatus int8       `gorm:"column:process_status;type:tinyint;not null;default:0;" json:"processStatus"`   // 管理状态：0 正常 1 封禁
	ViewCount     uint64     `gorm:"column:view_count;type:bigint unsigned;not null;default:0;" json:"viewCount"`   // 访问数量
	ReplyCount    uint64     `gorm:"column:reply_count;type:bigint unsigned;not null;default:0;" json:"replyCount"` // 访问数量
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`                  //
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"` //
}

func (itself *SmallEntity) TableName() string {
	return tableName
}
