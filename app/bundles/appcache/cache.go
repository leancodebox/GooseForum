package appcache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
)

const defaultTTL = 2 * time.Minute
const bytesPerMiB = 1024 * 1024
const maxEntrySizeBytes = 8 * 1024
const fallbackHardMaxCacheMB = 64
const minHardMaxCacheMB = 32
const maxHardMaxCacheMB = 512
const hardMaxMemoryPercent = 1

var cache *bigcache.BigCache

func init() {
	config := bigcache.DefaultConfig(defaultTTL)
	config.Shards = 32
	config.MaxEntriesInWindow = 4096
	config.MaxEntrySize = maxEntrySizeBytes
	config.HardMaxCacheSize = hardMaxCacheSizeMB()
	config.Verbose = false

	var err error
	cache, err = bigcache.New(context.Background(), config)
	if err != nil {
		slog.Error("appcache init failed", "err", err)
		return
	}
	closer.Register(cache.Close)
	slog.Info("appcache initialized",
		"ttl", defaultTTL,
		"hardMaxMB", config.HardMaxCacheSize,
		"maxEntryBytes", config.MaxEntrySize,
	)
}

func GetJSON[T any](key string) (T, bool) {
	var zero T
	if cache == nil {
		return zero, false
	}
	data, err := cache.Get(key)
	if err != nil {
		return zero, false
	}
	return jsonopt.Decode[T](data), true
}

func SetJSON[T any](key string, value T) error {
	if cache == nil {
		return errors.New("appcache unavailable")
	}
	data := []byte(jsonopt.Encode(value))
	if len(data) > maxEntrySizeBytes {
		slog.Debug("appcache: entry too large, skip store", "key", key, "size", len(data), "max", maxEntrySizeBytes)
		return nil
	}
	return cache.Set(key, data)
}

func GetOrLoadJSON[T any](key string, load func() (T, error)) T {
	if data, ok := GetJSON[T](key); ok {
		slog.Debug("appcache: hit", "key", key)
		return data
	}
	slog.Debug("appcache: miss", "key", key)
	res, err := load()
	if err != nil {
		slog.Debug("appcache: loader error", "key", key, "err", err)
		return res
	}
	if err = SetJSON(key, res); err != nil {
		slog.Debug("appcache: store error", "key", key, "err", err)
	}
	return res
}

func Delete(key string) {
	if cache == nil {
		return
	}
	_ = cache.Delete(key)
}

func hardMaxCacheSizeMB() int {
	return hardMaxCacheSizeMBFromMemory(totalMemoryBytes())
}

func hardMaxCacheSizeMBFromMemory(totalBytes uint64, ok bool) int {
	if !ok || totalBytes == 0 {
		return fallbackHardMaxCacheMB
	}

	calculatedMB := totalBytes * hardMaxMemoryPercent / 100 / bytesPerMiB
	if calculatedMB < minHardMaxCacheMB {
		return minHardMaxCacheMB
	}
	if calculatedMB > maxHardMaxCacheMB {
		return maxHardMaxCacheMB
	}
	return int(calculatedMB)
}
