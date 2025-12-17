package transform

import (
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

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
		ExternalInformation: user.ExternalInformation,
	}
}

func User2UserCard(user users.EntityComplete, stats userStatistics.Entity, isFollowing bool, isSelf bool, hasLogin bool, allowAvatarUpload bool) *vo.UserCard {
	return &vo.UserCard{
		UserId:            user.Id,
		Username:          user.Username,
		AvatarUrl:         user.GetWebAvatarUrl(),
		Bio:               user.Bio,
		Signature:         user.Signature,
		IsAdmin:           user.RoleId > 0,
		ArticleCount:      stats.ArticleCount,
		LikeReceivedCount: stats.LikeReceivedCount,
		FollowerCount:     stats.FollowerCount,
		IsOnline:          time.Since(stats.LastActiveTime) < 120*time.Second,
		IsFollowing:       isFollowing,
		ExternalInfo:      user.ExternalInformation,
		IsSelf:            isSelf,
		HasLogin:          hasLogin,
		AllowAvatarUpload: allowAvatarUpload,
		LastActiveTime:    stats.LastActiveTime,
	}
}
