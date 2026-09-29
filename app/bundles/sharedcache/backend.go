// Package sharedcache owns business caches whose backend can be replaced.
// The memory backend is process-local; it does not provide cluster coherence.
package sharedcache

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

// Backend operates within one cache namespace. Clear and DeletePrefix must not
// affect other namespaces. UpdateIfPresent must not create a missing entry.
// Implementations own serialization, expiration and backend-specific concurrency.
type Backend[V any] interface {
	Get(string) (V, bool, error)
	Set(string, V, time.Duration) error
	Delete(string) error
	DeletePrefix(string) error
	Clear() error
	UpdateIfPresent(string, func(V) V, time.Duration) (bool, error)
}

func configuredDriver() string {
	return strings.ToLower(strings.TrimSpace(preferences.GetString("cache.driver", "memory")))
}

type backendSelection struct {
	mu         sync.RWMutex
	driver     string
	configured bool
}

var selection = &backendSelection{driver: "memory"}

// ConfigureFromPreferences freezes the backend choice before serving requests.
// Failed configuration can be retried; later preference changes are ignored.
func ConfigureFromPreferences() error {
	return selection.configure(configuredDriver())
}

func (s *backendSelection) configure(driver string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.configured {
		return nil
	}
	if driver != "memory" {
		return fmt.Errorf("unsupported cache.driver %q (available: memory)", driver)
	}
	s.driver, s.configured = driver, true
	return nil
}

func newBackend[V any](name string, maxEntries uint64) (Backend[V], error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("sharedcache: cache name is required")
	}
	selection.mu.RLock()
	driver := selection.driver
	selection.mu.RUnlock()
	switch driver {
	case "memory":
		return &memoryBackend[V]{cache: localcache.Cache[V]{MaxEntries: maxEntries}}, nil
	default:
		return nil, fmt.Errorf("unsupported cache driver %q", driver)
	}
}

type memoryBackend[V any] struct {
	cache localcache.Cache[V]
}

func (b *memoryBackend[V]) Get(key string) (V, bool, error) {
	value, found := b.cache.Get(key)
	return value, found, nil
}

func (b *memoryBackend[V]) Set(key string, value V, ttl time.Duration) error {
	b.cache.Set(key, value, ttl)
	return nil
}

func (b *memoryBackend[V]) Delete(key string) error {
	b.cache.Delete(key)
	return nil
}

func (b *memoryBackend[V]) DeletePrefix(prefix string) error {
	b.cache.DeletePrefix(prefix)
	return nil
}

func (b *memoryBackend[V]) Clear() error {
	b.cache.Clear()
	return nil
}

func (b *memoryBackend[V]) UpdateIfPresent(key string, update func(V) V, ttl time.Duration) (bool, error) {
	return b.cache.UpdateIfPresent(key, update, ttl), nil
}
