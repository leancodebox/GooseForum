package unreadservice

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
)

const statusTTL = 2 * time.Minute

var statusCache = sharedcache.Cache[Status]{Name: "unread-status", MaxEntries: cacheconfig.Current().UnreadStatus}

type Status struct {
	Notifications          bool   `json:"notifications"`
	Messages               bool   `json:"messages"`
	LatestNotificationType string `json:"latestNotificationType,omitempty"`
}

func GetStatus(userID uint64) Status {
	if userID == 0 {
		return Status{}
	}
	return statusCache.GetOrLoad(cacheKey(userID), func() (Status, error) {
		return loadStatus(userID), nil
	}, statusTTL)
}

func GetStatusForAudience(userID uint64, readableCategoryIDs []uint64, filterAudience bool) Status {
	if userID == 0 {
		return Status{}
	}
	return statusCache.GetOrLoad(audienceCacheKey(userID, readableCategoryIDs, filterAudience), func() (Status, error) {
		return loadStatusForAudience(userID, readableCategoryIDs, filterAudience), nil
	}, statusTTL)
}

func Invalidate(userID uint64) {
	if userID == 0 {
		return
	}
	statusCache.Delete(cacheKey(userID))
	statusCache.DeletePrefix(cacheKey(userID) + ":audience:")
}

func loadStatus(userID uint64) Status {
	latest := eventNotification.GetLastUnread(userID)
	return statusFromLatest(userID, latest)
}

func loadStatusForAudience(userID uint64, readableCategoryIDs []uint64, filterAudience bool) Status {
	latest := eventNotification.GetLastUnreadForAudience(userID, readableCategoryIDs, filterAudience)
	return statusFromLatest(userID, latest)
}

func statusFromLatest(userID uint64, latest eventNotification.Entity) Status {
	return Status{
		Notifications:          latest.Id != 0,
		Messages:               imUserChatConfigs.HasUnread(userID),
		LatestNotificationType: latest.EventType,
	}
}

func cacheKey(userID uint64) string {
	return "user:unread:status:" + strconv.FormatUint(userID, 10)
}

func audienceCacheKey(userID uint64, readableCategoryIDs []uint64, filterAudience bool) string {
	key := cacheKey(userID) + ":audience:"
	if !filterAudience {
		return key + "all"
	}
	categoryIDs := append([]uint64(nil), readableCategoryIDs...)
	slices.Sort(categoryIDs)
	categoryIDs = slices.Compact(categoryIDs)
	var builder strings.Builder
	builder.WriteString(key)
	builder.WriteString("categories")
	for _, categoryID := range categoryIDs {
		if categoryID == 0 {
			continue
		}
		builder.WriteByte(':')
		builder.WriteString(strconv.FormatUint(categoryID, 10))
	}
	return builder.String()
}
