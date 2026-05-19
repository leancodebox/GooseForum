package datacache

import (
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Cache 本 cache 目前不包含定时清理，请在有限场景使用
type Cache[V any] struct {
	items sync.Map
	group singleflight.Group
}

type cacheItem[V any] struct {
	value      V
	expiration int64 // 秒时间戳
}

func (c *Cache[V]) GetOrLoad(
	key string,
	getData func() (V, error), // 数据加载函数
	timeout time.Duration, // 缓存超时时间
) (value V) {
	data, err := c.GetOrLoadE(key, getData, timeout)
	if err != nil {
		slog.Debug("datacache: load failed in GetOrLoad", "key", key, "err", err)
	}
	return data
}

// GetOrLoadE 核心调用方法（满足您要求的参数形式）
func (c *Cache[V]) GetOrLoadE(
	key string,
	getData func() (V, error), // 数据加载函数
	timeout time.Duration, // 缓存超时时间
) (V, error) {
	// 首次快速读取
	if val, ok := c.get(key); ok {
		slog.Debug("datacache: hit", "key", key)
		return val, nil
	}
	slog.Debug("datacache: miss", "key", key)

	// 使用 singleflight 确保同一个 key 只有一个 goroutine 执行加载
	// 将 key 转换为字符串作为 singleflight 的 key

	result, err, _ := c.group.Do(key, func() (any, error) {
		// 在 singleflight 内部再次检查缓存，防止在等待期间其他 goroutine 已加载
		if val, ok := c.get(key); ok {
			slog.Debug("datacache: hit after singleflight wait", "key", key)
			return val, nil
		}

		// 执行加载逻辑
		newVal, err := getData()
		if err != nil {
			slog.Debug("datacache: loader error", "key", key, "err", err)
			return *new(V), err
		}

		// 缓存结果（带超时控制）
		c.items.Store(key, &cacheItem[V]{
			value:      newVal,
			expiration: time.Now().Add(timeout).Unix(),
		})
		slog.Debug("datacache: stored", "key", key, "ttl", timeout)

		return newVal, nil
	})

	if err != nil {
		slog.Debug("datacache: load failed", "key", key, "err", err)
		return *new(V), err
	}

	return result.(V), nil
}

func (c *Cache[V]) Clear() {
	deleted := 0
	c.items.Range(func(key, value any) bool {
		c.items.Delete(key)
		deleted++
		return true
	})
	slog.Debug("datacache: cleared", "deleted", deleted)
}

// 私有方法：带过期检查的读取
func (c *Cache[V]) get(key string) (V, bool) {
	if item, ok := c.items.Load(key); ok {
		cached := item.(*cacheItem[V])
		if time.Now().Unix() < cached.expiration {
			return cached.value, true
		}
		c.items.Delete(key) // 自动清理过期项
		slog.Debug("datacache: expired", "key", key)
	}
	return *new(V), false
}
