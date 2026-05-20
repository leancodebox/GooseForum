package vo

import (
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
)

// UserCard is the compact public profile payload used by profile cards.
type UserCard struct {
	UserId            uint64                    `json:"userId"`
	Username          string                    `json:"username"`
	Nickname          string                    `json:"nickname"`
	AvatarUrl         string                    `json:"avatarUrl"`
	Bio               string                    `json:"bio"`
	Signature         string                    `json:"signature"`
	WebsiteName       string                    `json:"websiteName"`
	Website           string                    `json:"website"`
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
	Badges            []badgeservice.UserBadge  `json:"badges"`
	LastActiveTime    time.Time                 `json:"lastActiveTime"`
	CreatedAt         time.Time                 `json:"createdAt"`
}

// UserHoverCard is the minimal public profile payload used by hover cards.
type UserHoverCard struct {
	UserId            uint64                    `json:"userId"`
	Username          string                    `json:"username"`
	Nickname          string                    `json:"nickname"`
	AvatarUrl         string                    `json:"avatarUrl"`
	Bio               string                    `json:"bio"`
	Signature         string                    `json:"signature"`
	WebsiteName       string                    `json:"websiteName"`
	Website           string                    `json:"website"`
	ExternalInfo      users.ExternalInformation `json:"externalInformation"`
	IsAdmin           bool                      `json:"isAdmin"`
	ArticleCount      uint                      `json:"articleCount"`
	ReplyCount        uint                      `json:"replyCount"`
	LikeReceivedCount uint                      `json:"likeReceivedCount"`
	FollowerCount     uint                      `json:"followerCount"`
	IsOnline          bool                      `json:"isOnline"`
	IsFollowing       bool                      `json:"isFollowing"`
	Badges            []badgeservice.UserBadge  `json:"badges"`
	LastActiveTime    time.Time                 `json:"lastActiveTime"`
	CreatedAt         time.Time                 `json:"createdAt"`
}
