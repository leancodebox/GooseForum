package controllers

import (
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"sync"
	"time"
)

var articlesType = []datastruct.Option[string, int]{
	{Name: "分享", Value: 1},
	{Name: "求助", Value: 2},
}

var articlesTypeMap = array.Slice2Map(articlesType, func(v datastruct.Option[string, int]) int {
	return v.Value
})

var (
	siteStatsCacheHasCache bool
	siteStatsCache         SiteStats
	siteStatsCacheTime     time.Time
	siteStatsCacheMutex    sync.Mutex
)

type SiteStats struct {
	UserCount         int64 `json:"userCount"`
	UserMonthCount    int64 `json:"userMonthCount"`
	ArticleCount      int64 `json:"articleCount"`
	ArticleMonthCount int64 `json:"articleMonthCount"`
	Reply             int64 `json:"reply"`
}

func GetSiteStatisticsData() SiteStats {
	siteStatsCacheMutex.Lock()
	defer siteStatsCacheMutex.Unlock()

	if time.Since(siteStatsCacheTime) < 5*time.Second && siteStatsCacheHasCache {
		return siteStatsCache
	}

	result := SiteStats{
		UserCount:         users.GetCount(),
		UserMonthCount:    users.GetMonthCount(),
		ArticleCount:      articles.GetCount(),
		ArticleMonthCount: articles.GetMonthCount(),
		Reply:             reply.GetCount(),
	}

	siteStatsCache = result
	siteStatsCacheTime = time.Now()
	siteStatsCacheHasCache = true
	return siteStatsCache
}

// 初始化缓存
var articleCache = &Cache[string, []articles.SmallEntity]{}

func getRecommendedArticles() []articles.SmallEntity {
	data, _ := articleCache.GetOrLoad(
		"hot_articles",
		func() ([]articles.SmallEntity, error) {
			return articles.GetRecommendedArticles(4)
		},
		5*time.Minute, // 缓存5分钟
	)
	return data
}

func articlesSmallEntity2Dto(data []articles.SmallEntity) []ArticlesSimpleDto {
	userIds := array.Map(data, func(t articles.SmallEntity) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)

	//获取文章的分类信息
	articleIds := array.Map(data, func(t articles.SmallEntity) uint64 {
		return t.Id
	})
	categoryRs := articleCategoryRs.GetByArticleIdsEffective(articleIds)
	categoryIds := array.Map(categoryRs, func(t *articleCategoryRs.Entity) uint64 {
		return t.ArticleCategoryId
	})
	categoryMap := articleCategory.GetMapByIds(categoryIds)
	// 获取文章的分类和标签
	categoriesGroup := array.GroupBy(categoryRs, func(rs *articleCategoryRs.Entity) uint64 {
		return rs.ArticleId
	})
	return array.Map(data, func(t articles.SmallEntity) ArticlesSimpleDto {
		categoryNames := array.Map(categoriesGroup[t.Id], func(rs *articleCategoryRs.Entity) string {
			if category, ok := categoryMap[rs.ArticleCategoryId]; ok {
				return category.Category
			}
			return ""
		})
		username := ""
		avatarUrl := urlconfig.GetDefaultAvatar()
		if user, ok := userMap[t.UserId]; ok {
			username = user.Username
			avatarUrl = user.GetWebAvatarUrl()
		}
		return ArticlesSimpleDto{
			Id:             t.Id,
			Title:          t.Title,
			LastUpdateTime: t.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreateTime:     t.CreatedAt.Format("2006-01-02 15:04:05"),
			Username:       username,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			Category:       FirstOr(categoryNames, "未分类"),
			Categories:     categoryNames,
			CategoriesId: array.Map(categoriesGroup[t.Id], func(rs *articleCategoryRs.Entity) uint64 {
				return rs.ArticleCategoryId
			}),
			Type:    t.Type,
			TypeStr: articlesTypeMap[int(t.Type)].Name,
		}
	})
}
