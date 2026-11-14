package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

// Shared across all viewers. The existing cache coalesces concurrent misses
// and prevents invalidated loads from repopulating the cache.
var publicTopicMaxCache = &sharedcache.Cache[uint64]{Name: "public-topic-max", MaxEntries: 1}

func PublicTopicMaxID() (uint64, error) {
	return loadPublicTopicMax(topics.PublicMaxID)
}
func ClearPublicTopicMaxCache() { publicTopicMaxCache.Clear() }

func loadPublicTopicMax(load func() (uint64, error)) (uint64, error) {
	return publicTopicMaxCache.GetOrLoadE("max", load, 24*time.Hour)
}
