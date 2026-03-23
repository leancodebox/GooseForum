package vo

import (
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

type UserCard struct {
	UserId            uint64                    `json:"userId"`
	Username          string                    `json:"username"`
	Nickname          string                    `json:"nickname"`
	AvatarUrl         string                    `json:"avatarUrl"`
	Bio               string                    `json:"bio"`
	Signature         string                    `json:"signature"`
	IsAdmin           bool                      `json:"isAdmin"`
	ArticleCount      uint                      `json:"articleCount"`
	ReplyCount        uint                      `json:"replyCount"`
	LikeReceivedCount uint                      `json:"likeReceivedCount"`
	LikeGivenCount    uint                      `json:"likeGivenCount"`
	FollowerCount     uint                      `json:"followerCount"`
	FollowingCount    uint                      `json:"followingCount"`
	CollectionCount   uint                      `json:"collectionCount"`
	IsOnline          bool                      `json:"isOnline"`
	IsFollowing       bool                      `json:"isFollowing"`
	ExternalInfo      users.ExternalInformation `json:"externalInformation"`
	IsSelf            bool                      `json:"isSelf"`
	LastActiveTime    time.Time                 `json:"lastActiveTime"`
	CreatedAt         time.Time                 `json:"createdAt"`
}
