package usercardservice

import (
	"strconv"
	"sync/atomic"

	"github.com/leancodebox/GooseForum/app/bundles/appcache"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

// cardVersion scopes all user-card keys inside the shared app cache.
// Bumping it invalidates every card key without resetting unrelated cache entries.
var cardVersion atomic.Uint64

func GetHoverCard(userID uint64) (*vo.UserHoverCard, bool) {
	card, ok := getCached[vo.UserHoverCard](hoverKey(userID))
	if ok {
		return &card, true
	}
	user, stats, ok := loadUserStats(userID)
	if !ok {
		return nil, false
	}
	card = *transform.User2UserHoverCard(user, stats, false)
	setCached(hoverKey(userID), card)
	return &card, true
}

func GetCard(userID uint64) (*vo.UserCard, bool) {
	card, ok := getCached[vo.UserCard](cardKey(userID))
	if ok {
		return &card, true
	}
	user, stats, ok := loadUserStats(userID)
	if !ok {
		return nil, false
	}
	card = *transform.User2UserCard(user, stats, false, 0)
	setCached(cardKey(userID), card)
	return &card, true
}

func Invalidate(userID uint64) {
	if userID == 0 {
		return
	}
	appcache.Delete(hoverKey(userID))
	appcache.Delete(cardKey(userID))
}

func Clear() {
	// The underlying app cache is shared by unread status and user summaries, so
	// we cannot Reset it here. Versioned keys make old card entries unreachable
	// and let bigcache remove them naturally when the TTL expires.
	cardVersion.Add(1)
}

func loadUserStats(userID uint64) (users.EntityComplete, userStatistics.Entity, bool) {
	user, err := users.Get(userID)
	if err != nil || user.Id == 0 {
		return user, userStatistics.Entity{}, false
	}
	return user, userStatistics.Get(userID), true
}

func getCached[T any](key string) (T, bool) {
	return appcache.GetJSON[T](key)
}

func setCached[T any](key string, value T) {
	_ = appcache.SetJSON(key, value)
}

func hoverKey(userID uint64) string {
	return "user:card:v" + strconv.FormatUint(cardVersion.Load(), 10) + ":hover:" + strconv.FormatUint(userID, 10)
}

func cardKey(userID uint64) string {
	return "user:card:v" + strconv.FormatUint(cardVersion.Load(), 10) + ":full:" + strconv.FormatUint(userID, 10)
}
