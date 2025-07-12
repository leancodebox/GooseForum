package userStatistics

import (
	"time"
)

const tableName = "user_statistics"

// pid 用户ID
const pid = "user_id"

// fieldArticleCount 发表文章数
const fieldArticleCount = "article_count"

// fieldPostCount 发帖数(包括主题和回复)
const fieldPostCount = "post_count"

// fieldReplyCount 评论数
const fieldReplyCount = "reply_count"

// fieldFollowerCount 粉丝数
const fieldFollowerCount = "follower_count"

// fieldFollowingCount 关注数
const fieldFollowingCount = "following_count"

// fieldLikeReceivedCount 收到的点赞数
const fieldLikeReceivedCount = "like_received_count"

// fieldLikeGivenCount 给出的点赞数
const fieldLikeGivenCount = "like_given_count"

// fieldCollectionCount 收藏数
const fieldCollectionCount = "collection_count"

// fieldLastActiveTime 最后活跃时间
const fieldLastActiveTime = "last_active_time"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

type Entity struct {
	UserId            uint64     `gorm:"primaryKey;column:user_id;autoIncrement;not null;" json:"userId"`                           // 用户ID
	ArticleCount      uint       `gorm:"column:article_count;type:int unsigned;not null;default:0;" json:"articleCount"`            // 发表文章数
	ReplyCount        uint       `gorm:"column:reply_count;type:int unsigned;not null;default:0;" json:"replyCount"`                // 评论数
	FollowerCount     uint       `gorm:"column:follower_count;type:int unsigned;not null;default:0;" json:"followerCount"`          // 粉丝数
	FollowingCount    uint       `gorm:"column:following_count;type:int unsigned;not null;default:0;" json:"followingCount"`        // 关注数
	LikeReceivedCount uint       `gorm:"column:like_received_count;type:int unsigned;not null;default:0;" json:"likeReceivedCount"` // 收到的点赞数
	LikeGivenCount    uint       `gorm:"column:like_given_count;type:int unsigned;not null;default:0;" json:"likeGivenCount"`       // 给出的点赞数
	CollectionCount   uint       `gorm:"column:collection_count;type:int unsigned;not null;default:0;" json:"collectionCount"`      // 收藏数
	LastActiveTime    *time.Time `gorm:"column:last_active_time;type:datetime;" json:"lastActiveTime"`                              // 最后活跃时间
	CreatedAt         time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`                        //
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
