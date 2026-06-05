// Package transform maps model entities to API/view payload structs.
package transform

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

// User2userShow maps a user entity to the authenticated user summary payload.
func User2userShow(user users.EntityComplete) *vo.UserInfoShow {
	return &vo.UserInfoShow{
		UserId:              user.Id,
		Username:            user.Username,
		Email:               user.Email,
		Nickname:            user.Nickname,
		Bio:                 user.Bio,
		Signature:           user.Signature,
		Prestige:            user.Prestige,
		AvatarUrl:           user.GetWebAvatarUrl(),
		CreateTime:          user.CreatedAt,
		IsAdmin:             user.RoleId > 0,
		IsActivated:         user.IsActivated,
		ExternalInformation: user.ExternalInformation,
	}
}

// User2UserDetailedVo maps a user entity to the detailed profile payload.
func User2UserDetailedVo(user users.EntityComplete) *vo.UserDetailedVo {
	return &vo.UserDetailedVo{
		Id:                  user.Id,
		Username:            user.Username,
		Email:               user.Email,
		Nickname:            user.Nickname,
		AvatarUrl:           user.GetWebAvatarUrl(),
		Bio:                 user.Bio,
		Signature:           user.Signature,
		WebsiteName:         user.WebsiteName,
		Website:             user.Website,
		ExternalInformation: user.ExternalInformation,
		Prestige:            user.Prestige,
		CreatedAt:           user.CreatedAt,
	}
}
