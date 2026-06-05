package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/samber/lo"
)

const (
	categoryCacheTTL   = time.Minute
	finiteCacheEntries = 4
)

var articleCategoryCache = &localcache.Cache[articleCategorySnapshot]{MaxEntries: finiteCacheEntries}

type articleCategorySnapshot struct {
	list        []*articleCategory.Entity
	categoryMap map[uint64]*articleCategory.Entity
}

func GetArticleCategory() []*articleCategory.Entity {
	return loadArticleCategorySnapshot().list
}

// ArticleCategoryMap GetMapByIds 根据ID列表获取分类Map
func ArticleCategoryMap() map[uint64]*articleCategory.Entity {
	return loadArticleCategorySnapshot().categoryMap
}

func loadArticleCategorySnapshot() articleCategorySnapshot {
	data, _ := articleCategoryCache.GetOrLoadE(
		"articleCategorySnapshot",
		func() (articleCategorySnapshot, error) {
			list := articleCategory.All()
			return articleCategorySnapshot{
				list: list,
				categoryMap: lo.KeyBy(list, func(v *articleCategory.Entity) uint64 {
					return v.Id
				}),
			}, nil
		},
		categoryCacheTTL,
	)
	return data
}

func GetCategoryById(id uint64) *articleCategory.Entity {
	return ArticleCategoryMap()[id]
}

func ClearArticleCategoryCache() {
	articleCategoryCache.Clear()
	articleSimpleVoCache.Clear()
}
