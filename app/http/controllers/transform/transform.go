// Package transform maps model entities to API/view payload structs.
package transform

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
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
		CanAccessAdmin:      user.RoleId > 0,
		IsActivated:         user.IsActivated,
		ExternalInformation: user.ExternalInformation,
	}
}

// User2UserDetailedVo maps a user entity to the detailed profile payload.
func User2UserDetailedVo(user users.EntityComplete) *vo.UserDetailedVo {
	userBadges := badgeservice.GetUserBadges(user.Id)
	return &vo.UserDetailedVo{
		Id:                  user.Id,
		Username:            user.Username,
		Email:               user.Email,
		Nickname:            user.Nickname,
		AvatarUrl:           user.GetWebAvatarUrl(),
		ProfileCoverUrl:     user.ProfileCoverUrl,
		Bio:                 user.Bio,
		Signature:           user.Signature,
		WebsiteName:         user.WebsiteName,
		Website:             user.Website,
		Locale:              user.Locale,
		ExternalInformation: user.ExternalInformation,
		Prestige:            user.Prestige,
		WornBadgeCode:       user.WornBadgeCode,
		Badges:              userBadges,
		WearableBadges:      badgeservice.WearableBadgesFromList(userBadges),
		WornBadge:           badgeservice.WornBadgeFromList(userBadges, user.WornBadgeCode),
		CreatedAt:           user.CreatedAt,
	}
}
