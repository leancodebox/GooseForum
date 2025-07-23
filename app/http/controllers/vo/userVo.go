package vo

import (
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"time"
)

type UserInfoShow struct {
	UserId              uint64                    `json:"userId,omitempty"`
	Username            string                    `json:"username"`
	Bio                 string                    `json:"bio"`
	Signature           string                    `json:"Signature"`
	Prestige            int64                     `json:"prestige"`
	AvatarUrl           string                    `json:"avatarUrl"`
	UserPoint           int64                     `json:"userPoint"`
	CreateTime          time.Time                 `json:"createTime"`
	IsAdmin             bool                      `json:"isAdmin"`
	ExternalInformation users.ExternalInformation `json:"externalInformation"`
}
