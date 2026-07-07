package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/samber/lo"
)

const (
	categoryCacheTTL   = time.Minute
	finiteCacheEntries = 4
)

var categoryCache = &localcache.Cache[categorySnapshot]{MaxEntries: finiteCacheEntries}

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
	data, _ := categoryCache.GetOrLoadE(
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
	return data
}

func categoriesToArticleCategories(list []*category.Entity) []*articleCategory.Entity {
	res := make([]*articleCategory.Entity, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		res = append(res, &articleCategory.Entity{
			Id:        item.Id,
			Category:  item.Name,
			Desc:      item.Desc,
			Icon:      item.Icon,
			Color:     item.Color,
			Slug:      item.Slug,
			Sort:      item.Sort,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return res
}

func categoryMapToArticleCategoryMap(categoryMap map[uint64]*category.Entity) map[uint64]*articleCategory.Entity {
	res := make(map[uint64]*articleCategory.Entity, len(categoryMap))
	for id, item := range categoryMap {
		if item == nil {
			continue
		}
		res[id] = &articleCategory.Entity{
			Id:        item.Id,
			Category:  item.Name,
			Desc:      item.Desc,
			Icon:      item.Icon,
			Color:     item.Color,
			Slug:      item.Slug,
			Sort:      item.Sort,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}
	return res
}

func GetArticleCategory() []*articleCategory.Entity {
	return categoriesToArticleCategories(GetCategory())
}

// ArticleCategoryMap keeps legacy callers working while category callers migrate.
func ArticleCategoryMap() map[uint64]*articleCategory.Entity {
	return categoryMapToArticleCategoryMap(CategoryMap())
}

func GetCategoryById(id uint64) *articleCategory.Entity {
	return ArticleCategoryMap()[id]
}

func GetCleanCategoryById(id uint64) *category.Entity {
	return CategoryMap()[id]
}

func ClearCategoryCache() {
	categoryCache.Clear()
	ClearTopicListCache()
}

func ClearArticleCategoryCache() {
	ClearCategoryCache()
}
