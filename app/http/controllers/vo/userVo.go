package vo

import (
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

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
	IsAdmin             bool                      `json:"isAdmin"`
	ExternalInformation users.ExternalInformation `json:"externalInformation"`
}

type UserDetailedVo struct {
	Id                  uint64                    `json:"id"`
	Username            string                    `json:"username"`
	Email               string                    `json:"email"`
	Nickname            string                    `json:"nickname"`
	AvatarUrl           string                    `json:"avatarUrl"`
	Bio                 string                    `json:"bio"`
	Signature           string                    `json:"signature"`
	WebsiteName         string                    `json:"websiteName"`
	Website             string                    `json:"website"`
	ExternalInformation users.ExternalInformation `json:"externalInformation"`
	Prestige            int64                     `json:"prestige"`
	CreatedAt           time.Time                 `json:"createdAt"`
}
