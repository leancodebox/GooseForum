package authsessionservice

import (
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
)

func TestLocalAuthCacheInvalidationPreventsStaleRefill(t *testing.T) {
	cache := newLocalAuthCache()
	now := time.Now()
	session := authsessions.Token{Id: 1, UserId: 7, ExpiresAt: now.Add(time.Hour)}
	generation := cache.Generation()
	cache.InvalidateUser(7)
	if cache.Store("token", session, now, generation) {
		t.Fatal("stale lookup repopulated cache")
	}
	if !cache.Store("token", session, now, cache.Generation()) {
		t.Fatal("fresh lookup not cached")
	}
	if _, ok := cache.Get("token", now); !ok {
		t.Fatal("fresh session missing")
	}
	cache.InvalidateUser(7)
	if _, ok := cache.Get("token", now); ok {
		t.Fatal("revoked session remained cached")
	}
	generation = cache.Generation()
	cache.InvalidateUser(8)
	if !cache.Store("unrelated", session, now, generation) {
		t.Fatal("unrelated user invalidation blocked cache fill")
	}
	if !cache.Store("token", session, now, cache.Generation()) {
		t.Fatal("fresh lookup not cached")
	}
	if _, ok := cache.Get("token", now.Add(cacheTTL)); ok {
		t.Fatal("cache entry outlived TTL")
	}
}

func TestLocalAuthCacheEvictsOneEntryAtCapacity(t *testing.T) {
	cache := newLocalAuthCache()
	now := time.Now()
	for i := 0; i <= maxCacheEntries; i++ {
		key := fmt.Sprintf("session-%d", i)
		if !cache.Store(key, authsessions.Token{Id: uint64(i + 1), UserId: 7, ExpiresAt: now.Add(time.Hour)}, now, cache.Generation()) {
			t.Fatal("cache insert failed")
		}
	}
	if len(cache.entries) != maxCacheEntries || cache.order.Len() != maxCacheEntries {
		t.Fatalf("cache size = %d, order size = %d", len(cache.entries), cache.order.Len())
	}
	if _, ok := cache.Get("session-0", now); ok {
		t.Fatal("oldest entry was not evicted")
	}
	if _, ok := cache.Get(fmt.Sprintf("session-%d", maxCacheEntries), now); !ok {
		t.Fatal("newest entry missing")
	}
}

func TestLocalAuthCacheKeepsNewerSessionTimes(t *testing.T) {
	cache := newLocalAuthCache()
	now := time.Now()
	session := authsessions.Token{Id: 1, UserId: 7, ExpiresAt: now.Add(time.Hour), LastSeenAt: now}
	cache.Store("token", session, now, cache.Generation())
	newer := session
	newer.ExpiresAt = now.Add(2 * time.Hour)
	newer.LastSeenAt = now.Add(time.Minute)
	cache.UpdateIfPresent("token", newer)
	cache.UpdateIfPresent("token", session)
	got, ok := cache.Get("token", now)
	if !ok || !got.ExpiresAt.Equal(newer.ExpiresAt) || !got.LastSeenAt.Equal(newer.LastSeenAt) {
		t.Fatalf("older snapshot replaced newer session times: %+v", got)
	}
}
