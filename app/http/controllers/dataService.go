package controllers

import (
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"sync"
	"time"
)

var (
	cacheMutex   sync.RWMutex
	articleCache = struct {
		data       []articles.SmallEntity
		expiration time.Time
	}{
		expiration: time.Now().Add(-1 * time.Second), // 初始设为过期状态
	}
	cacheTTL = 5 * time.Minute // 缓存有效时间
)

func getRecommendedArticles() []articles.SmallEntity {
	// 先尝试读取缓存
	cacheMutex.RLock()
	if time.Now().Before(articleCache.expiration) {
		defer cacheMutex.RUnlock()
		return articleCache.data
	}
	cacheMutex.RUnlock()

	// 缓存过期，重新获取数据
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 双检锁避免并发重复更新
	if time.Now().Before(articleCache.expiration) {
		return articleCache.data
	}

	newData, _ := articles.GetRecommendedArticles(3)
	articleCache.data = newData
	articleCache.expiration = time.Now().Add(cacheTTL)

	return newData
}
