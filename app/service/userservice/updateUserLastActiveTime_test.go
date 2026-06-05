package userservice

import (
	"sync"
	"testing"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

func TestUserActivityStoreKeepsNewestTime(t *testing.T) {
	store := newTestActivityStore(t)
	userID := uint64(7)
	newer := time.Now()
	older := newer.Add(-time.Minute)

	store.remember(userID, newer)
	store.remember(userID, older)

	got, ok := store.get(userID)
	if !ok {
		t.Fatal("activity store did not remember user activity")
	}
	if !got.Equal(newer) {
		t.Fatalf("activity time = %v, want newest %v", got, newer)
	}
}

func TestUserActivityStoreFlushesOnExpiration(t *testing.T) {
	store := newTestActivityStore(t)
	userID := uint64(8)
	activeTime := time.Now()

	flushed := waitForFlush(t, store, func() {
		store.cache.Set(userID, userActivity{LastActiveAt: activeTime, LastFlushedAt: activeTime, Dirty: true}, 10*time.Millisecond)
		time.Sleep(30 * time.Millisecond)
		store.cache.DeleteExpired()
	})

	if flushed[userID] != activeTime {
		t.Fatalf("flushed activity = %v, want %v", flushed[userID], activeTime)
	}
}

func TestUserActivityStoreFlushesOnClose(t *testing.T) {
	store := newTestActivityStore(t)
	userID := uint64(9)
	activeTime := time.Now()

	flushed := waitForFlush(t, store, func() {
		store.remember(userID, activeTime)
		store.close()
	})

	if flushed[userID] != activeTime {
		t.Fatalf("flushed activity = %v, want %v", flushed[userID], activeTime)
	}
}

func TestUserActivityStoreFlushesWhenActivePastWindow(t *testing.T) {
	store := newTestActivityStore(t)
	userID := uint64(11)
	first := time.Now()
	soon := first.Add(userOnlineWindow - time.Second)
	later := first.Add(userOnlineWindow + time.Second)

	flushed := waitForFlush(t, store, func() {
		store.remember(userID, first)
		store.remember(userID, soon)
		store.remember(userID, later)
	})

	if flushed[userID] != later {
		t.Fatalf("flushed activity = %v, want %v", flushed[userID], later)
	}

	item := store.cache.Get(userID)
	if item == nil {
		t.Fatal("activity item missing after threshold flush")
	}
	if !item.Value().LastFlushedAt.Equal(later) {
		t.Fatalf("last flushed at = %v, want %v", item.Value().LastFlushedAt, later)
	}
}

func TestUserActivityStoreDoesNotFlushCleanActivityOnClose(t *testing.T) {
	store := newTestActivityStore(t)
	userID := uint64(12)
	first := time.Now()
	later := first.Add(userOnlineWindow + time.Second)

	flushCount := 0
	store.flushFn = func(uint64, time.Time) {
		flushCount++
	}
	store.remember(userID, first)
	store.remember(userID, later)
	store.close()

	if flushCount != 1 {
		t.Fatalf("flush count = %d, want 1", flushCount)
	}
}

func newTestActivityStore(t *testing.T) *userActivityStore {
	t.Helper()
	store := &userActivityStore{}
	store.flushFn = func(uint64, time.Time) {}
	store.init()
	t.Cleanup(func() {
		store.cache.Stop()
	})
	return store
}

func waitForFlush(t *testing.T, store *userActivityStore, run func()) map[uint64]time.Time {
	t.Helper()
	var mu sync.Mutex
	flushed := map[uint64]time.Time{}
	done := make(chan struct{}, 1)
	store.flushFn = func(userID uint64, activeTime time.Time) {
		mu.Lock()
		flushed[userID] = activeTime
		mu.Unlock()
		select {
		case done <- struct{}{}:
		default:
		}
	}

	run()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for activity flush")
	}

	mu.Lock()
	defer mu.Unlock()
	result := make(map[uint64]time.Time, len(flushed))
	for userID, activeTime := range flushed {
		result[userID] = activeTime
	}
	return result
}

func TestUserActivityStoreIgnoresDeletedEviction(t *testing.T) {
	store := newTestActivityStore(t)
	flushed := make(chan struct{}, 1)
	store.flushFn = func(uint64, time.Time) {
		flushed <- struct{}{}
	}
	now := time.Now()
	store.cache.Set(10, userActivity{LastActiveAt: now, LastFlushedAt: now, Dirty: true}, ttlcache.DefaultTTL)
	store.cache.Delete(10)
	select {
	case <-flushed:
		t.Fatal("deleted eviction should not flush activity")
	case <-time.After(30 * time.Millisecond):
	}
}
