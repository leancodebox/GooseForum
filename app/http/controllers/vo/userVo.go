package vo

import (
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
)

// UserInfoShow is the authenticated user summary exposed to pages and APIs.
type UserInfoShow struct {
	UserId              uint64                    `json:"userId,omitempty"`
	Username            string                    `json:"username"`
	Email               string                    `json:"email"`
	Nickname            string                    `json:"nickname"`
	Bio                 string                    `json:"bio"`
	Signature           string                    `json:"Signature"`
	Prestige            int64                     `json:"prestige"`
	AvatarUrl           string                    `json:"avatarUrl"`
	UserPoint           int64                     `json:"userPoint"`
	CreateTime          time.Time                 `json:"createTime"`
	CanAccessAdmin      bool                      `json:"canAccessAdmin"`
	IsActivated         int8                      `json:"isActivated"`
	ExternalInformation users.ExternalInformation `json:"externalInformation"`
}

// UserDetailedVo is the editable profile payload used by account settings.
type UserDetailedVo struct {
	Id                  uint64                    `json:"id"`
	Username            string                    `json:"username"`
	Email               string                    `json:"email"`
	Nickname            string                    `json:"nickname"`
	AvatarUrl           string                    `json:"avatarUrl"`
	ProfileCoverUrl     string                    `json:"profileCoverUrl"`
	Bio                 string                    `json:"bio"`
	Signature           string                    `json:"signature"`
	WebsiteName         string                    `json:"websiteName"`
	Website             string                    `json:"website"`
	Locale              string                    `json:"locale"`
	ExternalInformation users.ExternalInformation `json:"externalInformation"`
	Prestige            int64                     `json:"prestige"`
	WornBadgeCode       string                    `json:"wornBadgeCode"`
	Badges              []badgeservice.UserBadge  `json:"badges"`
	WearableBadges      []badgeservice.UserBadge  `json:"wearableBadges"`
	WornBadge           *badgeservice.UserBadge   `json:"wornBadge,omitempty"`
	CreatedAt           time.Time                 `json:"createdAt"`
}
