package topics

import (
	"time"

	"gorm.io/gorm"
)

const tableName = "topics"

type Entity struct {
	Id            uint64         `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_topics_list_default,priority:5,sort:desc;index:idx_topics_list_hot,priority:4,sort:desc;index:idx_topics_list_popular,priority:4,sort:desc;index:idx_topics_list_new,priority:4,sort:desc;index:idx_topics_admin_list,priority:3,sort:desc;index:idx_topics_admin_user_list,priority:4,sort:desc;" json:"id"`
	Title         string         `gorm:"column:title;type:varchar(512);not null;default:'';" json:"title"`
	CategoryIds   []uint64       `gorm:"column:category_id;type:varchar(255);not null;default:'[]';serializer:json" json:"categoryIds"`
	UserId        uint64         `gorm:"column:user_id;type:bigint unsigned;not null;default:0;index:idx_topics_user_status,priority:1;index:idx_topics_admin_user_list,priority:1;" json:"userId"`
	Status        int8           `gorm:"column:status;type:tinyint;not null;default:0;index:idx_topics_user_status,priority:2;index:idx_topics_list_default,priority:1;index:idx_topics_list_hot,priority:1;index:idx_topics_list_popular,priority:1;index:idx_topics_list_new,priority:1;" json:"status"`
	ProcessStatus int8           `gorm:"column:process_status;type:tinyint;not null;default:0;index:idx_topics_user_status,priority:3;index:idx_topics_list_default,priority:2;index:idx_topics_list_hot,priority:2;index:idx_topics_list_popular,priority:2;index:idx_topics_list_new,priority:2;" json:"processStatus"`
	LikeCount     uint64         `gorm:"column:like_count;type:bigint unsigned;not null;default:0;" json:"likeCount"`
	ViewCount     uint64         `gorm:"column:view_count;type:bigint unsigned;not null;default:0;index:idx_topics_list_popular,priority:3,sort:desc;" json:"viewCount"`
	PostCount     uint64         `gorm:"column:post_count;type:bigint unsigned;not null;default:0;" json:"postCount"`
	ReplyCount    uint64         `gorm:"column:reply_count;type:bigint unsigned;not null;default:0;index:idx_topics_list_hot,priority:3,sort:desc;" json:"replyCount"`
	PostSeq       uint64         `gorm:"column:post_seq;type:bigint unsigned;not null;default:0;" json:"postSeq"`
	FirstPostId   uint64         `gorm:"column:first_post_id;type:bigint unsigned;not null;default:0;" json:"firstPostId"`
	LastPostId    uint64         `gorm:"column:last_post_id;type:bigint unsigned;not null;default:0;" json:"lastPostId"`
	LastPostedAt  *time.Time     `gorm:"column:last_posted_at;index:idx_topics_last_posted,sort:desc;" json:"lastPostedAt"`
	PinWeight     int            `gorm:"column:pin_weight;type:int;not null;default:0;index:idx_topics_list_default,priority:3,sort:desc;index:idx_topics_admin_list,priority:1,sort:desc;index:idx_topics_admin_user_list,priority:2,sort:desc;" json:"pinWeight"`
	Excerpt       string         `gorm:"column:excerpt;type:varchar(255);not null;default:'';" json:"excerpt"`
	FirstImageURL string         `gorm:"column:first_image_url;type:varchar(512);not null;default:'';" json:"firstImageUrl"`
	Posters       []Poster       `gorm:"column:posters;type:text;serializer:json" json:"posters"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime;<-:create;index:idx_topics_list_new,priority:3,sort:desc;" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime;index:idx_topics_list_default,priority:4,sort:desc;index:idx_topics_admin_list,priority:2,sort:desc;index:idx_topics_admin_user_list,priority:3,sort:desc;" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-"`
}

type Poster struct {
	UserID      uint64 `json:"user_id"`
	Description string `json:"description"`
}

func (itself *Entity) TableName() string {
	return tableName
}

func (itself *Entity) GetPosters() []Poster {
	if len(itself.Posters) == 0 {
		return []Poster{{UserID: itself.UserId}}
	}
	return itself.Posters
}
