package vo

import (
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

type UserCard struct {
	UserId            uint64                    `json:"userId"`
	Username          string                    `json:"username"`
	AvatarUrl         string                    `json:"avatarUrl"`
	Bio               string                    `json:"bio"`
	Signature         string                    `json:"signature"`
	IsAdmin           bool                      `json:"isAdmin"`
	ArticleCount      uint                      `json:"articleCount"`
	LikeReceivedCount uint                      `json:"likeReceivedCount"`
	FollowerCount     uint                      `json:"followerCount"`
	IsOnline          bool                      `json:"isOnline"`
	IsFollowing       bool                      `json:"isFollowing"`
	ExternalInfo      users.ExternalInformation `json:"externalInformation"`
	IsSelf            bool                      `json:"isSelf"`
	HasLogin          bool                      `json:"hasLogin"`
	AllowAvatarUpload bool                      `json:"allowAvatarUpload"`
	LastActiveTime    time.Time                 `json:"lastActiveTime"`
}
