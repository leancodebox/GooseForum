package unreadservice

import (
	"context"
	"strconv"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
)

const statusTTL = 30 * time.Second

type Status struct {
	Notifications          bool   `json:"notifications"`
	Messages               bool   `json:"messages"`
	LatestNotificationType string `json:"latestNotificationType,omitempty"`
}

var statusCache = newStatusCache()

func newStatusCache() *bigcache.BigCache {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(statusTTL))
	if err != nil {
		return nil
	}
	return cache
}

func GetStatus(userID uint64) Status {
	if userID == 0 {
		return Status{}
	}
	key := cacheKey(userID)
	if statusCache != nil {
		if data, err := statusCache.Get(key); err == nil {
			return jsonopt.Decode[Status](data)
		}
	}

	status := loadStatus(userID)
	if statusCache != nil {
		_ = statusCache.Set(key, []byte(jsonopt.Encode(status)))
	}
	return status
}

func Invalidate(userID uint64) {
	if userID == 0 || statusCache == nil {
		return
	}
	_ = statusCache.Delete(cacheKey(userID))
}

func loadStatus(userID uint64) Status {
	latest := eventNotification.GetLastUnread(userID)
	return Status{
		Notifications:          latest.Id != 0,
		Messages:               imUserChatConfigs.HasUnread(userID),
		LatestNotificationType: latest.EventType,
	}
}

func cacheKey(userID uint64) string {
	return "unread-status:" + strconv.FormatUint(userID, 10)
}
