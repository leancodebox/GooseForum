package unreadservice

import (
	"strconv"

	"github.com/leancodebox/GooseForum/app/bundles/appcache"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
)

type Status struct {
	Notifications          bool   `json:"notifications"`
	Messages               bool   `json:"messages"`
	LatestNotificationType string `json:"latestNotificationType,omitempty"`
}

func GetStatus(userID uint64) Status {
	if userID == 0 {
		return Status{}
	}
	key := cacheKey(userID)
	if status, ok := appcache.GetJSON[Status](key); ok {
		return status
	}

	status := loadStatus(userID)
	_ = appcache.SetJSON(key, status)
	return status
}

func Invalidate(userID uint64) {
	if userID == 0 {
		return
	}
	appcache.Delete(cacheKey(userID))
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
	return "user:unread:status:" + strconv.FormatUint(userID, 10)
}
