package sharedcache

import (
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Cache must not be copied after first use. Name is a stable namespace for a
// future shared backend. Each memory-backed Cache retains its own entries.
type Cache[V any] struct {
	Name       string
	MaxEntries uint64

	backend    Backend[V]
	group      singleflight.Group
	mu         sync.Mutex
	flights    map[string]*flight
	nextFlight uint64
}

type flight struct {
	id    uint64
	users int
}

type loadResult[V any] struct{ value V }

// NewWithBackend injects a namespace-scoped backend, including in tests.
func NewWithBackend[V any](name string, backend Backend[V]) *Cache[V] {
	if name == "" || backend == nil {
		panic("sharedcache: name and backend are required")
	}
	return &Cache[V]{Name: name, backend: backend}
}

func (c *Cache[V]) init() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.backend == nil {
		backend, err := newBackend[V](c.Name, c.MaxEntries)
		if err != nil {
			return err
		}
		c.backend = backend
	}
	return nil
}

func (c *Cache[V]) GetOrLoad(key string, load func() (V, error), ttl time.Duration) V {
	value, err := c.GetOrLoadE(key, load, ttl)
	c.logError("load", err)
	return value
}

func (c *Cache[V]) GetOrLoadE(key string, load func() (V, error), ttl time.Duration) (V, error) {
	if err := c.init(); err != nil {
		return *new(V), err
	}
	c.mu.Lock()
	value, found, err := c.backend.Get(key)
	if err != nil || found {
		c.mu.Unlock()
		return value, err
	}
	if c.flights == nil {
		c.flights = make(map[string]*flight)
	}
	pending := c.flights[key]
	if pending == nil {
		c.nextFlight++
		pending = &flight{id: c.nextFlight}
		c.flights[key] = pending
	}
	pending.users++
	c.mu.Unlock()
	defer c.releaseFlight(key, pending)
	flightKey := strconv.FormatUint(pending.id, 10)
	result, err, _ := c.group.Do(flightKey, func() (any, error) {
		c.mu.Lock()
		value, found, err := c.backend.Get(key)
		c.mu.Unlock()
		if err != nil || found {
			return loadResult[V]{value}, err
		}
		value, err = load()
		if err != nil {
			return loadResult[V]{}, err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		// An invalidation or explicit write during loading makes this fill stale.
		if c.flights[key] == pending {
			if err := c.backend.Set(key, value, ttl); err != nil {
				c.logError("store", err)
			}
		}
		return loadResult[V]{value}, nil
	})
	if err != nil {
		return *new(V), err
	}
	return result.(loadResult[V]).value, nil
}

func (c *Cache[V]) releaseFlight(key string, pending *flight) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pending.users--
	if pending.users == 0 && c.flights[key] == pending {
		delete(c.flights, key)
	}
}

func (c *Cache[V]) Set(key string, value V, ttl time.Duration) error {
	if err := c.init(); err != nil {
		c.logError("set", err)
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.flights, key)
	err := c.backend.Set(key, value, ttl)
	c.logError("set", err)
	return err
}

func (c *Cache[V]) Delete(key string) error {
	if err := c.init(); err != nil {
		c.logError("delete", err)
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.flights, key)
	err := c.backend.Delete(key)
	c.logError("delete", err)
	return err
}

func (c *Cache[V]) DeletePrefix(prefix string) error {
	if err := c.init(); err != nil {
		c.logError("delete prefix", err)
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.flights {
		if strings.HasPrefix(key, prefix) {
			delete(c.flights, key)
		}
	}
	err := c.backend.DeletePrefix(prefix)
	c.logError("delete prefix", err)
	return err
}

func (c *Cache[V]) Clear() error {
	if err := c.init(); err != nil {
		c.logError("clear", err)
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flights = nil
	err := c.backend.Clear()
	c.logError("clear", err)
	return err
}

func (c *Cache[V]) UpdateIfPresent(key string, update func(V) V, ttl time.Duration) bool {
	if err := c.init(); err != nil {
		c.logError("update", err)
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// The database may already have changed even if no cached entry exists.
	delete(c.flights, key)
	updated, err := c.backend.UpdateIfPresent(key, update, ttl)
	c.logError("update", err)
	return updated && err == nil
}

func (c *Cache[V]) logError(operation string, err error) {
	if err != nil {
		slog.Error("sharedcache operation failed", "cache", c.Name, "operation", operation, "err", err)
	}
}
