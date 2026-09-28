package authsessionservice

import (
	"container/list"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
)

const (
	cacheTTL        = 15 * time.Second
	maxCacheEntries = 10000
)

type authCache interface {
	Get(string, time.Time) (authsessions.Token, bool)
	Generation() uint64
	Store(string, authsessions.Token, time.Time, uint64) bool
	UpdateIfPresent(string, authsessions.Token)
	InvalidateUser(uint64)
}

var cachedSessions authCache = newLocalAuthCache()

type cachedToken struct {
	session authsessions.Token
	until   time.Time
	order   *list.Element
}

type localAuthCache struct {
	mu             sync.RWMutex
	generation     uint64
	entries        map[string]cachedToken
	order          *list.List
	userGeneration map[uint64]uint64
}

func newLocalAuthCache() *localAuthCache {
	return &localAuthCache{entries: make(map[string]cachedToken), order: list.New(), userGeneration: make(map[uint64]uint64)}
}

func (c *localAuthCache) Get(key string, now time.Time) (authsessions.Token, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || !entry.until.After(now) || !entry.session.ExpiresAt.After(now) {
		return authsessions.Token{}, false
	}
	return entry.session, true
}

func (c *localAuthCache) Generation() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.generation
}

func (c *localAuthCache) Store(key string, session authsessions.Token, now time.Time, generation uint64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.userGeneration[session.UserId] > generation {
		return false
	}
	if old, ok := c.entries[key]; ok {
		c.order.Remove(old.order)
	} else if len(c.entries) >= maxCacheEntries {
		oldest := c.order.Front()
		delete(c.entries, oldest.Value.(string))
		c.order.Remove(oldest)
	}
	c.entries[key] = cachedToken{session: session, until: now.Add(cacheTTL), order: c.order.PushBack(key)}
	return true
}

func (c *localAuthCache) UpdateIfPresent(key string, session authsessions.Token) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.entries[key]; ok && entry.session.Id == session.Id {
		if session.ExpiresAt.After(entry.session.ExpiresAt) {
			entry.session.ExpiresAt = session.ExpiresAt
		}
		if session.LastSeenAt.After(entry.session.LastSeenAt) {
			entry.session.LastSeenAt = session.LastSeenAt
		}
		c.entries[key] = entry
	}
}

func (c *localAuthCache) InvalidateUser(userID uint64) {
	c.mu.Lock()
	for key, entry := range c.entries {
		if entry.session.UserId == userID {
			c.order.Remove(entry.order)
			delete(c.entries, key)
		}
	}
	c.generation++
	c.userGeneration[userID] = c.generation
	c.mu.Unlock()
}
