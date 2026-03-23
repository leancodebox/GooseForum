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

func User2UserCard(user users.EntityComplete, stats userStatistics.Entity, isFollowing bool, currentUserId uint64) *vo.UserCard {
	return &vo.UserCard{
		UserId:            user.Id,
		Username:          user.Username,
		Nickname:          user.Nickname,
		AvatarUrl:         user.GetWebAvatarUrl(),
		Bio:               user.Bio,
		Signature:         user.Signature,
		IsAdmin:           user.RoleId > 0,
		ArticleCount:      stats.ArticleCount,
		ReplyCount:        stats.ReplyCount,
		LikeReceivedCount: stats.LikeReceivedCount,
		LikeGivenCount:    stats.LikeGivenCount,
		FollowerCount:     stats.FollowerCount,
		FollowingCount:    stats.FollowingCount,
		CollectionCount:   stats.CollectionCount,
		IsOnline:          time.Since(stats.LastActiveTime) < 120*time.Second,
		IsFollowing:       isFollowing,
		ExternalInfo:      user.ExternalInformation,
		IsSelf:            currentUserId == user.Id,
		LastActiveTime:    stats.LastActiveTime,
		CreatedAt:         user.CreatedAt,
	}
}

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
