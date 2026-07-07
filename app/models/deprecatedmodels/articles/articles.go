package articles

import (
	"fmt"
	"time"

	"gorm.io/gorm"
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
const fieldReplyCount = "reply_count"

// fieldReplySeq 回复序列号
const fieldReplySeq = "reply_seq"

// fieldViewCount 浏览量
const fieldViewCount = "view_count"

// fieldPinWeight 全站置顶权重，0 表示不置顶，数字越大越靠前
const fieldPinWeight = "pin_weight"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id              uint64         `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_user_published_id,priority:4,sort:desc;index:idx_articles_list_default,priority:5,sort:desc;index:idx_articles_list_hot,priority:4,sort:desc;index:idx_articles_list_popular,priority:4,sort:desc;index:idx_articles_list_new,priority:4,sort:desc;" json:"id"` //
	Title           string         `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"`                                                                                                                                                                                                                                                    //
	Content         string         `gorm:"column:content;type:text;" json:"content"`                                                                                                                                                                                                                                                                            //
	Description     string         `gorm:"column:description;type:varchar(255);not null;default:'';" json:"description"`                                                                                                                                                                                                                                        // 文章描述，用于SEO
	FirstImageURL   string         `gorm:"column:first_image_url;type:varchar(512);not null;default:'';" json:"firstImageUrl"`                                                                                                                                                                                                                                  // 正文首图，用于SEO和分享
	RenderedHTML    string         `gorm:"column:rendered_html;type:text;" json:"renderedHTML"`                                                                                                                                                                                                                                                                 //md 渲染后数据
	RenderedVersion uint32         `gorm:"column:rendered_version;type:bigint unsigned;not null;default:0;" json:"renderedVersion"`                                                                                                                                                                                                                             //md 的渲染器版本
	Type            int8           `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`                                                                                                                                                                                                                                                            // 文章类型：0 博文，1教程，2问答，3分享
	CategoryId      []uint64       `gorm:"column:category_id;type:varchar(255);not null;default:'[]';serializer:json" json:"categoryId"`                                                                                                                                                                                                                        // 分类
	UserId          uint64         `gorm:"column:user_id;type:bigint unsigned;not null;default:0;index:idx_user_status;index:idx_user_published_id,priority:1;" json:"userId"`                                                                                                                                                                                  //
	Posters         []Poster       `gorm:"column:posters;type:text;serializer:json" json:"posters"`
	ArticleStatus   int8           `gorm:"column:article_status;type:tinyint;not null;default:0;index:idx_user_status;index:idx_user_published_id,priority:2;index:idx_articles_list_default,priority:1;index:idx_articles_list_hot,priority:1;index:idx_articles_list_popular,priority:1;index:idx_articles_list_new,priority:1;" json:"articleStatus"` // 文章状态：0 草稿 1 发布
	ProcessStatus   int8           `gorm:"column:process_status;type:tinyint;not null;default:0;index:idx_user_status;index:idx_user_published_id,priority:3;index:idx_articles_list_default,priority:2;index:idx_articles_list_hot,priority:2;index:idx_articles_list_popular,priority:2;index:idx_articles_list_new,priority:2;" json:"processStatus"` // 管理状态：0 正常 1 封禁
	LikeCount       uint64         `gorm:"column:like_count;type:bigint unsigned;not null;default:0;" json:"likeCount"`                                                                                                                                                                                                                                  // 喜欢数量
	ViewCount       uint64         `gorm:"column:view_count;type:bigint unsigned;not null;default:0;index:idx_articles_list_popular,priority:3,sort:desc;" json:"viewCount"`                                                                                                                                                                             // 访问数量
	ReplyCount      uint64         `gorm:"column:reply_count;type:bigint unsigned;not null;default:0;index:idx_articles_list_hot,priority:3,sort:desc;" json:"replyCount"`                                                                                                                                                                               // 评论数量
	ReplySeq        uint64         `gorm:"column:reply_seq;type:bigint unsigned;not null;default:0;" json:"replySeq"`                                                                                                                                                                                                                                    // 历史最大回复序号
	PinWeight       int            `gorm:"column:pin_weight;type:int;not null;default:0;index:idx_articles_list_default,priority:3,sort:desc;" json:"pinWeight"`                                                                                                                                                                                         // 全站置顶权重，0 表示不置顶
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime;<-:create;index:idx_articles_list_new,priority:3,sort:desc;" json:"createdAt"`                                                                                                                                                                                                //
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime;index:idx_articles_list_default,priority:4,sort:desc;" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt //
}

type Type int

const (
	Share Type = iota + 1
	Help
)

// Poster stores one participant avatar entry.
type Poster struct {
	UserID      uint64 `json:"user_id"`
	Description string `json:"description"`
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
	FirstImageURL string         `gorm:"column:first_image_url;type:varchar(512);not null;default:'';" json:"firstImageUrl"`           // 正文首图，用于SEO和分享
	Type          int8           `gorm:"column:type;type:tinyint;not null;default:0;" json:"type"`                                     // 文章类型：0 博文，1教程，2问答，3分享
	CategoryId    []uint64       `gorm:"column:category_id;type:varchar(255);not null;default:'[]';serializer:json" json:"categoryId"` // 分类
	UserId        uint64         `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`                        //
	Posters       []Poster       `gorm:"column:posters;type:text;serializer:json" json:"posters"`
	ArticleStatus int8           `gorm:"column:article_status;type:tinyint;not null;default:0;" json:"articleStatus"`   // 文章状态：0 草稿 1 发布
	ProcessStatus int8           `gorm:"column:process_status;type:tinyint;not null;default:0;" json:"processStatus"`   // 管理状态：0 正常 1 封禁
	ViewCount     uint64         `gorm:"column:view_count;type:bigint unsigned;not null;default:0;" json:"viewCount"`   // 访问数量
	ReplyCount    uint64         `gorm:"column:reply_count;type:bigint unsigned;not null;default:0;" json:"replyCount"` // 评论
	ReplySeq      uint64         `gorm:"column:reply_seq;type:bigint unsigned;not null;default:0;" json:"replySeq"`     // 历史最大回复序号
	LikeCount     uint64         `gorm:"column:like_count;type:bigint unsigned;not null;default:0;" json:"likeCount"`   // 被喜欢
	PinWeight     int            `gorm:"column:pin_weight;type:int;not null;default:0;" json:"pinWeight"`               // 全站置顶权重，0 表示不置顶
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`                  //
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt //
}

func (itself *SmallEntity) TableName() string {
	return tableName
}

func (itself *SmallEntity) PubDate() string {
	now := time.Now()
	duration := now.Sub(itself.CreatedAt)

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
	}

	return itself.CreatedAt.Format("2006-01-02")
}

func (itself *SmallEntity) GetPosters() []Poster {
	if len(itself.Posters) == 0 {
		return []Poster{{UserID: itself.UserId}}
	}
	return itself.Posters
}
