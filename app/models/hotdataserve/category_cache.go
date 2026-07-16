package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/samber/lo"
)

const (
	categoryCacheTTL = time.Minute
)

var categoryCache = &localcache.Cache[categorySnapshot]{MaxEntries: cacheconfig.Current().Category}

type categorySnapshot struct {
	list        []*category.Entity
	categoryMap map[uint64]*category.Entity
}

func GetCategory() []*category.Entity {
	return loadCategorySnapshot().list
}

func CategoryMap() map[uint64]*category.Entity {
	return loadCategorySnapshot().categoryMap
}

func loadCategorySnapshot() categorySnapshot {
	return categoryCache.GetOrLoad(
		"categorySnapshot",
		func() (categorySnapshot, error) {
			list := category.All()
			return categorySnapshot{
				list: list,
				categoryMap: lo.KeyBy(list, func(v *category.Entity) uint64 {
					return v.Id
				}),
			}, nil
		},
		categoryCacheTTL,
	)
}

func GetCleanCategoryById(id uint64) *category.Entity {
	return CategoryMap()[id]
}

func ClearCategoryCache() {
	categoryCache.Clear()
	ClearTopicListCache()
}

func ClearTopicCategoryCache() {
	ClearCategoryCache()
}
