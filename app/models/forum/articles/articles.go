package articles

import (
	"fmt"
	"gorm.io/gorm"
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
	Id              uint64         `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                            //
	Title           string         `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"`                                  //
	Content         string         `gorm:"column:content;type:text;" json:"content"`                                                          //
	Description     string         `gorm:"column:description;type:varchar(255);not null;default:'';" json:"description"`                      // 文章描述，用于SEO
	RenderedHTML    string         `gorm:"column:rendered_html;type:text;" json:"renderedHTML"`                                               //md 渲染后数据
	RenderedVersion uint32         `gorm:"column:rendered_version;type:bigint unsigned;not null;default:0;" json:"renderedVersion"`           //md 的渲染器版本
	Type            int8           `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`                                          // 文章类型：0 博文，1教程，2问答，3分享
	CategoryId      []uint64       `gorm:"column:category_id;type:varchar(255);not null;default:'[]';serializer:json" json:"categoryId"`      // 分类
	UserId          uint64         `gorm:"column:user_id;type:bigint unsigned;not null;default:0;index:idx_user_status;" json:"userId"`       //
	ArticleStatus   int8           `gorm:"column:article_status;type:tinyint;not null;default:0;index:idx_user_status;" json:"articleStatus"` // 文章状态：0 草稿 1 发布
	ProcessStatus   int8           `gorm:"column:process_status;type:tinyint;not null;default:0;index:idx_user_status;" json:"processStatus"` // 管理状态：0 正常 1 封禁
	LikeCount       uint64         `gorm:"column:like_count;type:bigint unsigned;not null;default:0;" json:"likeCount"`                       // 喜欢数量
	ViewCount       uint64         `gorm:"column:view_count;index;type:bigint unsigned;not null;default:0;" json:"viewCount"`                 // 访问数量
	ReplyCount      uint64         `gorm:"column:reply_count;type:bigint unsigned;not null;default:0;" json:"replyCount"`                     // 评论数量
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`                                      //
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime;index;" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt //
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
	Id            uint64         `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                       //
	Title         string         `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"`                             //
	Description   string         `gorm:"column:description;type:varchar(255);not null;default:'';" json:"description"`                 // 文章描述，用于SEO
	Type          int8           `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`                                     // 文章类型：0 博文，1教程，2问答，3分享
	CategoryId    []uint64       `gorm:"column:category_id;type:varchar(255);not null;default:'[]';serializer:json" json:"categoryId"` // 分类
	UserId        uint64         `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`                        //
	ArticleStatus int8           `gorm:"column:article_status;type:tinyint;not null;default:0;" json:"articleStatus"`                  // 文章状态：0 草稿 1 发布
	ProcessStatus int8           `gorm:"column:process_status;type:tinyint;not null;default:0;" json:"processStatus"`                  // 管理状态：0 正常 1 封禁
	ViewCount     uint64         `gorm:"column:view_count;type:bigint unsigned;not null;default:0;" json:"viewCount"`                  // 访问数量
	ReplyCount    uint64         `gorm:"column:reply_count;type:bigint unsigned;not null;default:0;" json:"replyCount"`                // 评论
	LikeCount     uint64         `gorm:"column:like_count;type:bigint unsigned;not null;default:0;" json:"likeCount"`                  // 被喜欢
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`                                 //
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt //
}

func (itself *SmallEntity) TableName() string {
	return tableName
}

func (itself *SmallEntity) PubDate() string {
	now := time.Now()
	duration := now.Sub(itself.CreatedAt)

	// 使用秒数进行计算，避免重复的浮点运算
	seconds := int64(duration.Seconds())

	if seconds < 60 {
		return "刚刚"
	} else if seconds < 3600 { // 60 * 60
		minutes := seconds / 60
		return fmt.Sprintf("%d分钟前", minutes)
	} else if seconds < 86400 { // 24 * 60 * 60
		hours := seconds / 3600
		return fmt.Sprintf("%d小时前", hours)
	} else if seconds < 604800 { // 7 * 24 * 60 * 60
		days := seconds / 86400
		return fmt.Sprintf("%d天前", days)
	} else {
		// 超过7天显示具体日期
		return itself.CreatedAt.Format("2006-01-02")
	}
}
