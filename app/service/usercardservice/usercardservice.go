package usercardservice

import (
	"context"
	"strconv"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

const cardTTL = 2 * time.Minute

var cardCache = newCardCache()

func newCardCache() *bigcache.BigCache {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(cardTTL))
	if err != nil {
		return nil
	}
	return cache
}

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
	if userID == 0 || cardCache == nil {
		return
	}
	_ = cardCache.Delete(hoverKey(userID))
	_ = cardCache.Delete(cardKey(userID))
}

func Clear() {
	if cardCache != nil {
		_ = cardCache.Reset()
	}
}

func loadUserStats(userID uint64) (users.EntityComplete, userStatistics.Entity, bool) {
	user, err := users.Get(userID)
	if err != nil || user.Id == 0 {
		return user, userStatistics.Entity{}, false
	}
	return user, userStatistics.Get(userID), true
}

func getCached[T any](key string) (T, bool) {
	var zero T
	if cardCache == nil {
		return zero, false
	}
	data, err := cardCache.Get(key)
	if err != nil {
		return zero, false
	}
	return jsonopt.Decode[T](data), true
}

func setCached[T any](key string, value T) {
	if cardCache == nil {
		return
	}
	_ = cardCache.Set(key, []byte(jsonopt.Encode(value)))
}

func hoverKey(userID uint64) string {
	return "user-hover-card:" + strconv.FormatUint(userID, 10)
}

func cardKey(userID uint64) string {
	return "user-card:" + strconv.FormatUint(userID, 10)
}
