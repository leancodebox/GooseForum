package controllers

import "github.com/leancodebox/GooseForum/app/models/forum/users"

func user2userShow(user users.EntityComplete) UserInfoShow {
	return UserInfoShow{
		UserId:              user.Id,
		Username:            user.Username,
		Bio:                 user.Bio,
		Signature:           user.Signature,
		Prestige:            user.Prestige,
		AvatarUrl:           user.GetWebAvatarUrl(),
		CreateTime:          user.CreatedAt,
		IsAdmin:             user.RoleId > 0,
		ExternalInformation: user.ExternalInformation,
	}
}
