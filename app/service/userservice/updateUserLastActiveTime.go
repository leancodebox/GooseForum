package userservice

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
)

type userActivity struct {
	LastActiveAt  time.Time
	LastFlushedAt time.Time
	Dirty         bool
}

type userActivityStore struct {
	once           sync.Once
	cache          *ttlcache.Cache[uint64, userActivity]
	flushFn        func(uint64, time.Time)
	registerCloser bool
	closeMu        sync.RWMutex
	closed         bool
}

var activityStore = &userActivityStore{
	flushFn:        flushUserActivity,
	registerCloser: true,
}

func (store *userActivityStore) init() {
	store.once.Do(func() {
		store.cache = ttlcache.New[uint64, userActivity](
			ttlcache.WithCapacity[uint64, userActivity](cacheconfig.Current().UserActivity),
			ttlcache.WithDisableTouchOnHit[uint64, userActivity](),
		)
		store.cache.OnEviction(func(_ context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[uint64, userActivity]) {
			if reason == ttlcache.EvictionReasonDeleted {
				return
			}
			if store.isClosed() {
				return
			}
			store.flushPending(item.Key(), item.Value())
		})
		go store.cache.Start()
		if store.registerCloser {
			closer.RegisterPriority(closer.PriorityFlush, CloseUpdateUserLastActiveTime)
		}
	})
}

func (store *userActivityStore) remember(userID uint64, activeTime time.Time) {
	if userID == 0 {
		return
	}
	if activeTime.IsZero() {
		activeTime = time.Now()
	}
	store.init()
	store.closeMu.Lock()
	defer store.closeMu.Unlock()
	if store.closed {
		return
	}
	activity := userActivity{
		LastActiveAt:  activeTime,
		LastFlushedAt: activeTime,
		Dirty:         true,
	}
	if item := store.cache.Get(userID); item != nil {
		activity = item.Value()
		if activity.LastActiveAt.After(activeTime) {
			return
		}
		activity.LastActiveAt = activeTime
		activity.Dirty = true
		if shouldFlushUserActivity(activity) {
			store.flush(userID, activity.LastActiveAt)
			activity.LastFlushedAt = activity.LastActiveAt
			activity.Dirty = false
		}
	}
	store.cache.Set(userID, activity, userOnlineWindow)
}

func (store *userActivityStore) get(userID uint64) (time.Time, bool) {
	if userID == 0 {
		return time.Time{}, false
	}
	store.init()
	store.closeMu.RLock()
	defer store.closeMu.RUnlock()
	if store.closed {
		return time.Time{}, false
	}
	item := store.cache.Get(userID)
	if item == nil {
		return time.Time{}, false
	}
	return item.Value().LastActiveAt, true
}

func (store *userActivityStore) flush(userID uint64, activeTime time.Time) {
	if userID == 0 || activeTime.IsZero() || store.flushFn == nil {
		return
	}
	store.flushFn(userID, activeTime)
	touchUserPublicProfileActivity(userID, activeTime)
}

func (store *userActivityStore) close() {
	store.init()
	store.closeMu.Lock()
	if store.closed {
		store.closeMu.Unlock()
		return
	}
	store.closed = true
	pending := make(map[uint64]userActivity)
	store.cache.Range(func(item *ttlcache.Item[uint64, userActivity]) bool {
		pending[item.Key()] = item.Value()
		return true
	})
	store.closeMu.Unlock()

	store.cache.Stop()

	for userID, activity := range pending {
		store.flushPending(userID, activity)
	}
}

func (store *userActivityStore) flushPending(userID uint64, activity userActivity) {
	if !activity.Dirty {
		return
	}
	store.flush(userID, activity.LastActiveAt)
}

func shouldFlushUserActivity(activity userActivity) bool {
	return !activity.LastFlushedAt.IsZero() && activity.LastActiveAt.Sub(activity.LastFlushedAt) >= userOnlineWindow
}

func (store *userActivityStore) isClosed() bool {
	store.closeMu.RLock()
	closed := store.closed
	store.closeMu.RUnlock()
	return closed
}

func UpdateUserActivityAt(userID uint64, activeTime time.Time) {
	rememberUserActivity(userID, activeTime)
}

// CloseUpdateUserLastActiveTime stops the global activity cache and flushes current activity.
func CloseUpdateUserLastActiveTime() error {
	activityStore.close()
	return nil
}

func rememberUserActivity(userID uint64, activeTime time.Time) {
	activityStore.remember(userID, activeTime)
}

func recentUserActivity(userID uint64) (time.Time, bool) {
	return activityStore.get(userID)
}

func flushUserActivity(userID uint64, activeTime time.Time) {
	if rows := userStatistics.UpdateUserActivity(userID, activeTime); rows == 0 {
		slog.Debug("user activity flush did not update a row", "userID", userID)
	}
}
