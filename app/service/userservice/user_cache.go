package userservice

import (
	"errors"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

const (
	userInfoTTL          = 2 * time.Minute
	userPublicProfileTTL = 2 * time.Minute
	userOnlineWindow     = 120 * time.Second
)

var errUserNotFound = errors.New("user not found")

// UserInfo is the sanitized user snapshot cached by userservice.
// It intentionally excludes password hashes and model-only deletion metadata.
type UserInfo struct {
	Id                  uint64
	Username            string
	Email               string
	TokenVersion        uint64
	IsFrozen            int8
	IsActivated         int8
	ActivatedAt         *time.Time
	Nickname            string
	RoleId              uint64
	Prestige            int64
	AvatarUrl           string
	ProfileCoverUrl     string
	Bio                 string
	Signature           string
	WebsiteName         string
	Website             string
	ExternalInformation users.ExternalInformation
	WornBadgeCode       string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// UserPublicInfo contains only fields needed by public display surfaces.
type UserPublicInfo struct {
	Id                  uint64
	Username            string
	Nickname            string
	IsFrozen            int8
	RoleId              uint64
	Prestige            int64
	AvatarUrl           string
	ProfileCoverUrl     string
	Bio                 string
	Signature           string
	WebsiteName         string
	Website             string
	ExternalInformation users.ExternalInformation
	WornBadgeCode       string
	CreatedAt           time.Time
}

type UserPublicProfile struct {
	User      UserPublicInfo
	Stats     userStatistics.Entity
	Badges    []badgeservice.UserBadge
	WornBadge *badgeservice.UserBadge
}

var (
	userInfoCache          = localcache.Cache[UserInfo]{MaxEntries: 2048}
	userPublicProfileCache = localcache.Cache[UserPublicProfile]{MaxEntries: 1024}
)

func GetUserInfo(userID uint64) (UserInfo, bool) {
	if userID == 0 {
		return UserInfo{}, false
	}
	user, err := userInfoCache.GetOrLoadE(userInfoKey(userID), func() (UserInfo, error) {
		entity, err := users.Get(userID)
		if err != nil || entity.Id == 0 {
			return UserInfo{}, errUserNotFound
		}
		return userInfoFromEntity(entity), nil
	}, userInfoTTL)
	if err != nil {
		return UserInfo{}, false
	}
	return user, true
}

func GetUserRoleId(userID uint64) (uint64, bool) {
	user, ok := GetUserInfo(userID)
	if !ok {
		return 0, false
	}
	return user.RoleId, true
}

func GetUserHoverCard(userID uint64) (*vo.UserHoverCard, bool) {
	profile, ok := GetUserPublicProfile(userID)
	if !ok {
		return nil, false
	}
	card := buildUserHoverCard(profile)
	return &card, true
}

func GetUserCard(userID uint64) (*vo.UserCard, bool) {
	profile, ok := GetUserPublicProfile(userID)
	if !ok {
		return nil, false
	}
	card := buildUserCard(profile)
	return &card, true
}

func GetUserShow(userID uint64) *vo.UserInfoShow {
	if userID == 0 {
		return &vo.UserInfoShow{}
	}
	user, ok := GetUserInfo(userID)
	if !ok {
		return &vo.UserInfoShow{}
	}
	return transform.User2userShow(user.toEntity())
}

func GetUserPublicProfile(userID uint64) (UserPublicProfile, bool) {
	if userID == 0 {
		return UserPublicProfile{}, false
	}
	profile, err := userPublicProfileCache.GetOrLoadE(userPublicProfileKey(userID), func() (UserPublicProfile, error) {
		user, ok := GetUserInfo(userID)
		if !ok {
			return UserPublicProfile{}, errUserNotFound
		}
		badges := badgeservice.GetUserBadges(userID)
		return UserPublicProfile{
			User:      user.toPublicInfo(),
			Stats:     userStatistics.Get(userID),
			Badges:    badges,
			WornBadge: badgeservice.WornBadgeFromList(badges, user.WornBadgeCode),
		}, nil
	}, userPublicProfileTTL)
	if err != nil {
		return UserPublicProfile{}, false
	}
	profile.Badges = cloneUserBadges(profile.Badges)
	profile.WornBadge = cloneUserBadgePtr(profile.WornBadge)
	return profile, true
}

func InvalidateUserPublicProfileCache(userID uint64) {
	if userID == 0 {
		return
	}
	userPublicProfileCache.Delete(userPublicProfileKey(userID))
}

func ClearUserPublicProfileCache() {
	userPublicProfileCache.Clear()
}

func touchUserPublicProfileActivity(userID uint64, lastActiveTime time.Time) {
	if userID == 0 || lastActiveTime.IsZero() {
		return
	}
	userPublicProfileCache.UpdateIfPresent(userPublicProfileKey(userID), func(profile UserPublicProfile) UserPublicProfile {
		if lastActiveTime.After(profile.Stats.LastActiveTime) {
			profile.Stats.LastActiveTime = lastActiveTime
		}
		return profile
	}, userPublicProfileTTL)
}

func effectiveLastActiveTime(userID uint64, stored time.Time) time.Time {
	recent, ok := recentUserActivity(userID)
	if ok && recent.After(stored) {
		return recent
	}
	return stored
}

func refreshUserInfo(user users.EntityComplete) {
	if user.Id == 0 {
		return
	}
	info := userInfoFromEntity(user)
	userInfoCache.Set(userInfoKey(user.Id), info, userInfoTTL)
	InvalidateUserPublicProfileCache(user.Id)
}

func userInfoFromEntity(user users.EntityComplete) UserInfo {
	return UserInfo{
		Id:                  user.Id,
		Username:            user.Username,
		Email:               user.Email,
		TokenVersion:        user.TokenVersion,
		IsFrozen:            user.IsFrozen,
		IsActivated:         user.IsActivated,
		ActivatedAt:         user.ActivatedAt,
		Nickname:            user.Nickname,
		RoleId:              user.RoleId,
		Prestige:            user.Prestige,
		AvatarUrl:           user.AvatarUrl,
		ProfileCoverUrl:     user.ProfileCoverUrl,
		Bio:                 user.Bio,
		Signature:           user.Signature,
		WebsiteName:         user.WebsiteName,
		Website:             user.Website,
		ExternalInformation: user.ExternalInformation,
		WornBadgeCode:       user.WornBadgeCode,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}
}

func (user UserInfo) toPublicInfo() UserPublicInfo {
	return UserPublicInfo{
		Id:                  user.Id,
		Username:            user.Username,
		Nickname:            user.Nickname,
		IsFrozen:            user.IsFrozen,
		RoleId:              user.RoleId,
		Prestige:            user.Prestige,
		AvatarUrl:           user.AvatarUrl,
		ProfileCoverUrl:     user.ProfileCoverUrl,
		Bio:                 user.Bio,
		Signature:           user.Signature,
		WebsiteName:         user.WebsiteName,
		Website:             user.Website,
		ExternalInformation: user.ExternalInformation,
		WornBadgeCode:       user.WornBadgeCode,
		CreatedAt:           user.CreatedAt,
	}
}

func (user UserInfo) toEntity() users.EntityComplete {
	return users.EntityComplete{
		Id:                  user.Id,
		Username:            user.Username,
		Email:               user.Email,
		TokenVersion:        user.TokenVersion,
		IsFrozen:            user.IsFrozen,
		IsActivated:         user.IsActivated,
		ActivatedAt:         user.ActivatedAt,
		Nickname:            user.Nickname,
		RoleId:              user.RoleId,
		Prestige:            user.Prestige,
		AvatarUrl:           user.AvatarUrl,
		ProfileCoverUrl:     user.ProfileCoverUrl,
		Bio:                 user.Bio,
		Signature:           user.Signature,
		WebsiteName:         user.WebsiteName,
		Website:             user.Website,
		ExternalInformation: user.ExternalInformation,
		WornBadgeCode:       user.WornBadgeCode,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}
}

func (user UserPublicInfo) webAvatarURL() string {
	entity := users.EntityComplete{
		Id:        user.Id,
		Username:  user.Username,
		Nickname:  user.Nickname,
		IsFrozen:  user.IsFrozen,
		AvatarUrl: user.AvatarUrl,
	}
	return entity.GetWebAvatarUrl()
}

func buildUserCard(profile UserPublicProfile) vo.UserCard {
	user := profile.User
	stats := profile.Stats
	lastActiveTime := effectiveLastActiveTime(user.Id, stats.LastActiveTime)
	return vo.UserCard{
		UserId:            user.Id,
		Username:          user.Username,
		Nickname:          user.Nickname,
		AvatarUrl:         user.webAvatarURL(),
		ProfileCoverUrl:   user.ProfileCoverUrl,
		Bio:               user.Bio,
		Signature:         user.Signature,
		WebsiteName:       user.WebsiteName,
		Website:           user.Website,
		Prestige:          user.Prestige,
		ExternalInfo:      user.ExternalInformation,
		IsAdmin:           userIsAdmin(user.RoleId),
		ArticleCount:      stats.ArticleCount,
		ReplyCount:        stats.ReplyCount,
		LikeReceivedCount: stats.LikeReceivedCount,
		LikeGivenCount:    stats.LikeGivenCount,
		FollowerCount:     stats.FollowerCount,
		FollowingCount:    stats.FollowingCount,
		CollectionCount:   stats.CollectionCount,
		IsOnline:          time.Since(lastActiveTime) < userOnlineWindow,
		IsFollowing:       false,
		IsSelf:            false,
		Badges:            cloneUserBadges(profile.Badges),
		WornBadge:         cloneUserBadgePtr(profile.WornBadge),
		LastActiveTime:    lastActiveTime,
		CreatedAt:         user.CreatedAt,
	}
}

func buildUserHoverCard(profile UserPublicProfile) vo.UserHoverCard {
	user := profile.User
	stats := profile.Stats
	lastActiveTime := effectiveLastActiveTime(user.Id, stats.LastActiveTime)
	return vo.UserHoverCard{
		UserId:            user.Id,
		Username:          user.Username,
		Nickname:          user.Nickname,
		AvatarUrl:         user.webAvatarURL(),
		ProfileCoverUrl:   user.ProfileCoverUrl,
		Bio:               user.Bio,
		Signature:         user.Signature,
		WebsiteName:       user.WebsiteName,
		Website:           user.Website,
		Prestige:          user.Prestige,
		ExternalInfo:      user.ExternalInformation,
		IsAdmin:           userIsAdmin(user.RoleId),
		ArticleCount:      stats.ArticleCount,
		ReplyCount:        stats.ReplyCount,
		LikeReceivedCount: stats.LikeReceivedCount,
		FollowerCount:     stats.FollowerCount,
		IsOnline:          time.Since(lastActiveTime) < userOnlineWindow,
		IsFollowing:       false,
		Badges:            cloneUserBadges(profile.Badges),
		WornBadge:         cloneUserBadgePtr(profile.WornBadge),
		LastActiveTime:    lastActiveTime,
		CreatedAt:         user.CreatedAt,
	}
}

func userIsAdmin(roleID uint64) bool {
	return roleID > 0 && permission.CheckRole(roleID, permission.Admin)
}

func cloneUserBadges(items []badgeservice.UserBadge) []badgeservice.UserBadge {
	if len(items) == 0 {
		return []badgeservice.UserBadge{}
	}
	result := make([]badgeservice.UserBadge, len(items))
	copy(result, items)
	return result
}

func SetWornBadge(userID uint64, badgeCode string) bool {
	if userID == 0 {
		return false
	}
	if badgeCode != "" && badgeservice.GetWornBadge(userID, badgeCode) == nil {
		return false
	}
	if err := users.UpdateWornBadgeCode(userID, badgeCode); err != nil {
		return false
	}
	userInfoCache.UpdateIfPresent(userInfoKey(userID), func(user UserInfo) UserInfo {
		user.WornBadgeCode = badgeCode
		return user
	}, userInfoTTL)
	InvalidateUserPublicProfileCache(userID)
	return true
}

func cloneUserBadgePtr(item *badgeservice.UserBadge) *badgeservice.UserBadge {
	if item == nil {
		return nil
	}
	clone := *item
	return &clone
}

func userInfoKey(userID uint64) string {
	return "user:info:" + strconv.FormatUint(userID, 10)
}

func userPublicProfileKey(userID uint64) string {
	return "user:public-profile:" + strconv.FormatUint(userID, 10)
}
