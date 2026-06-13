package localcache

import (
	"log/slog"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"golang.org/x/sync/singleflight"
)

// defaultMaxEntries is the memory backstop for request/user-derived key spaces.
// Finite caches should still be documented and invalidated explicitly.
const defaultMaxEntries uint64 = 2048

// Cache is a small in-process cache facade backed by ttlcache.
type Cache[V any] struct {
	MaxEntries uint64

	once  sync.Once
	cache *ttlcache.Cache[string, V]
	group singleflight.Group
}

func (c *Cache[V]) init() {
	c.once.Do(func() {
		c.cache = ttlcache.New[string, V](
			ttlcache.WithCapacity[string, V](c.maxEntries()),
			ttlcache.WithDisableTouchOnHit[string, V](),
		)
		go c.cache.Start()
		closer.RegisterPriority(closer.PriorityCache, func() error {
			c.cache.Stop()
			return nil
		})
	})
}

func (c *Cache[V]) maxEntries() uint64 {
	if c.MaxEntries == 0 {
		return defaultMaxEntries
	}
	return c.MaxEntries
}

func (c *Cache[V]) GetOrLoad(
	key string,
	getData func() (V, error),
	timeout time.Duration,
) (value V) {
	data, err := c.GetOrLoadE(key, getData, timeout)
	if err != nil {
		slog.Debug("localcache: load failed in GetOrLoad", "key", key, "err", err)
	}
	return data
}

func (c *Cache[V]) GetOrLoadE(
	key string,
	getData func() (V, error),
	timeout time.Duration,
) (V, error) {
	c.init()
	if item := c.cache.Get(key); item != nil {
		slog.Debug("localcache: hit", "key", key)
		return item.Value(), nil
	}
	slog.Debug("localcache: miss", "key", key)

	result, err, _ := c.group.Do(key, func() (any, error) {
		if item := c.cache.Get(key); item != nil {
			slog.Debug("localcache: hit after singleflight wait", "key", key)
			return item.Value(), nil
		}
		newVal, err := getData()
		if err != nil {
			slog.Debug("localcache: loader error", "key", key, "err", err)
			return *new(V), err
		}
		c.cache.Set(key, newVal, timeout)
		slog.Debug("localcache: stored", "key", key, "ttl", timeout)
		return newVal, nil
	})
	if err != nil {
		slog.Debug("localcache: load failed", "key", key, "err", err)
		return *new(V), err
	}
	return result.(V), nil
}

func (c *Cache[V]) Clear() {
	c.init()
	c.cache.DeleteAll()
}

func (c *Cache[V]) Set(key string, value V, timeout time.Duration) {
	c.init()
	c.cache.Set(key, value, timeout)
	slog.Debug("localcache: set", "key", key, "ttl", timeout)
}

func (c *Cache[V]) UpdateIfPresent(key string, update func(V) V, timeout time.Duration) bool {
	c.init()
	item := c.cache.Get(key)
	if item == nil {
		return false
	}
	c.cache.Set(key, update(item.Value()), timeout)
	slog.Debug("localcache: updated", "key", key, "ttl", timeout)
	return true
}

func (c *Cache[V]) Delete(key string) {
	c.init()
	c.cache.Delete(key)
	slog.Debug("localcache: deleted", "key", key)
}
