package controllers

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	items sync.Map
	mu    sync.RWMutex
}

type cacheItem[V any] struct {
	value      V
	expiration int64 // 纳秒时间戳
}

// GetOrLoad 核心调用方法（满足您要求的参数形式）
func (c *Cache[K, V]) GetOrLoad(
	key K,
	getData func() (V, error), // 数据加载函数
	timeout time.Duration, // 缓存超时时间
) (V, error) {
	// 首次快速读取
	if val, ok := c.get(key); ok {
		return val, nil
	}

	// 加锁防止并发加载
	c.mu.Lock()
	defer c.mu.Unlock()

	// 二次检查防止重复加载
	if val, ok := c.get(key); ok {
		return val, nil
	}

	// 执行加载逻辑
	newVal, err := getData()
	if err != nil {
		return *new(V), err
	}

	// 缓存结果（带超时控制）
	c.items.Store(key, cacheItem[V]{
		value:      newVal,
		expiration: time.Now().Add(timeout).UnixNano(),
	})

	return newVal, nil
}

// 私有方法：带过期检查的读取
func (c *Cache[K, V]) get(key K) (V, bool) {
	if item, ok := c.items.Load(key); ok {
		cached := item.(cacheItem[V])
		if time.Now().UnixNano() < cached.expiration {
			return cached.value, true
		}
		c.items.Delete(key) // 自动清理过期项
	}
	return *new(V), false
}
