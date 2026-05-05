package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/datacache"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// 初始化缓存
var (
	siteStatisticsDataCache = &datacache.Cache[*vo.SiteStats]{}
	articleCache            = &datacache.Cache[[]*articles.SmallEntity]{}
	articleSimpleVoCache    = &datacache.Cache[[]*vo.ArticlesSimpleVo]{}
	articleCategoryCache    = &datacache.Cache[[]*articleCategory.Entity]{}
	articleCategoryMapCache = &datacache.Cache[map[uint64]*articleCategory.Entity]{}
)

var articlesType = []datastruct.Option[string, int]{
	{Name: "分享", Value: int(articles.Share)},
	{Name: "求助", Value: int(articles.Help)},
}

var articlesTypeMap = lo.KeyBy(articlesType, func(v datastruct.Option[string, int]) int {
	return v.Value
})

func GetArticlesType() *[]datastruct.Option[string, int] {
	return &articlesType
}

func GetArticlesTypeName(iType int) string {
	return articlesTypeMap[iType].Name
}

func GetLatestArticleSimpleVo() []*vo.ArticlesSimpleVo {
	return articleSimpleVoCache.GetOrLoad("home:GetLatestArticles", func() ([]*vo.ArticlesSimpleVo, error) {
		res := ArticlesSmallEntity2Vo(GetLatestArticles())
		return res, nil
	}, time.Second*10)
}

func GetLatestArticlesSimpleVoPaginated(page int, sort string) []*vo.ArticlesSimpleVo {
	if page < 1 {
		page = 1
	}
	// Default sort to latest if empty
	if sort == "" {
		sort = "latest"
	}
	key := "home:GetLatestArticles:" + sort + ":" + cast.ToString(page)
	return articleSimpleVoCache.GetOrLoad(key, func() ([]*vo.ArticlesSimpleVo, error) {
		res := articles.Page[articles.SmallEntity](articles.PageQuery{
			Page:         page,
			PageSize:     20,
			FilterStatus: true,
			Sort:         sort,
		})
		return ArticlesSmallEntity2Vo(lo.Map(res.Data, func(item articles.SmallEntity, _ int) *articles.SmallEntity {
			return &item
		})), nil
	}, time.Second*10)
}

func GetArticlesByCategorySimpleVo(categoryId uint64, sort string, page int) []*vo.ArticlesSimpleVo {
	if page < 1 {
		page = 1
	}
	key := "GetArticlesByCategory:" + cast.ToString(categoryId) + ":" + sort + ":" + cast.ToString(page)
	return articleSimpleVoCache.GetOrLoad(key, func() ([]*vo.ArticlesSimpleVo, error) {
		res := articles.Page[articles.SmallEntity](articles.PageQuery{
			Page:         page,
			PageSize:     20,
			Categories:   []int{int(categoryId)},
			FilterStatus: true,
			Sort:         sort,
		})
		return ArticlesSmallEntity2Vo(lo.Map(res.Data, func(item articles.SmallEntity, _ int) *articles.SmallEntity {
			return &item
		})), nil
	}, time.Second*10)
}

func GetSiteStatisticsData() *vo.SiteStats {
	data, _ := siteStatisticsDataCache.GetOrLoadE("", func() (*vo.SiteStats, error) {
		res := GetFriendLinksConfigCache()
		linksCount := lo.SumBy(res, func(group pageConfig.FriendLinksGroup) int {
			return len(group.Links)
		})
		return &vo.SiteStats{
			UserCount:         users.GetMaxId(),
			UserMonthCount:    dailyStats.GetCurrentMonthSum(dailyStats.StatTypeRegCount),
			ArticleCount:      articles.GetMaxId(),
			ArticleMonthCount: dailyStats.GetCurrentMonthSum(dailyStats.StatTypeArticleCount),
			Reply:             reply.GetMaxId(),
			LinksCount:        linksCount,
		}, nil
	}, time.Second*5)
	return data
}

func GetRecommendedArticles() []*articles.SmallEntity {
	data, _ := articleCache.GetOrLoadE(
		"GetRecommendedArticles",
		func() ([]*articles.SmallEntity, error) {
			res, err := articles.GetRecommendedArticles(4)
			return lo.Map(res, func(item articles.SmallEntity, _ int) *articles.SmallEntity {
				return &item
			}), err
		},
		5*time.Minute,
	)
	return data
}

func GetLatestArticles() []*articles.SmallEntity {
	data, _ := articleCache.GetOrLoadE(
		"GetLatestArticles",
		func() ([]*articles.SmallEntity, error) {
			res, err := articles.GetLatestArticles(20)
			return lo.Map(res, func(item articles.SmallEntity, _ int) *articles.SmallEntity {
				return &item
			}), err
		},
		10*time.Second,
	)
	return data
}

func GetArticleCategory() []*articleCategory.Entity {
	data, _ := articleCategoryCache.GetOrLoadE(
		"GetArticleCategory",
		func() ([]*articleCategory.Entity, error) {
			return articleCategory.All(), nil

		},
		1*time.Minute,
	)
	return data
}

func ArticleCategoryLabel() []datastruct.Option[string, uint64] {
	return lo.Map(GetArticleCategory(), func(t *articleCategory.Entity, _ int) datastruct.Option[string, uint64] {
		return datastruct.Option[string, uint64]{
			Name:  t.Category,
			Value: t.Id,
		}
	})
}

// ArticleCategoryMap GetMapByIds 根据ID列表获取分类Map
func ArticleCategoryMap() map[uint64]*articleCategory.Entity {
	data, _ := articleCategoryMapCache.GetOrLoadE(
		"GetArticleCategory",
		func() (map[uint64]*articleCategory.Entity, error) {
			return lo.KeyBy(articleCategory.All(), func(v *articleCategory.Entity) uint64 {
				return v.Id
			}), nil
		},
		1*time.Minute, // 缓存5分钟
	)
	return data
}

func GetCategoryById(id uint64) *articleCategory.Entity {
	return ArticleCategoryMap()[id]
}

func ArticlesSmallEntity2Vo(data []*articles.SmallEntity) []*vo.ArticlesSimpleVo {
	userIds := lo.Map(data, func(t *articles.SmallEntity, _ int) uint64 {
		return t.UserId
	})
	// Collect all user IDs from posters
	posterUserIds := lo.FlatMap(data, func(article *articles.SmallEntity, _ int) []uint64 {
		return lo.Map(article.GetPosters(), func(poster articles.Poster, _ int) uint64 {
			return poster.UserID
		})
	})
	userIds = append(userIds, posterUserIds...)
	userMap := users.GetMapByIds(lo.Uniq(userIds))
	return ArticlesSmallEntityWithUser2Vo(data, userMap)
}

func ArticlesSmallEntityWithUser2Vo(data []*articles.SmallEntity, userMap map[uint64]*users.EntityComplete) []*vo.ArticlesSimpleVo {
	categoryMap := ArticleCategoryMap()
	return lo.Map(data, func(t *articles.SmallEntity, _ int) *vo.ArticlesSimpleVo {
		categoryNames := lo.Map(t.CategoryId, func(item uint64, _ int) string {
			if category, ok := categoryMap[item]; ok && category != nil {
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

		// Map posters to Vo
		postersVo := lo.Map(t.GetPosters(), func(poster articles.Poster, _ int) vo.PosterVo {
			posterUsername := ""
			posterAvatarUrl := urlconfig.GetDefaultAvatar()
			if user, ok := userMap[poster.UserID]; ok {
				posterUsername = user.Username
				posterAvatarUrl = user.GetWebAvatarUrl()
			}
			return vo.PosterVo{
				Id:        poster.UserID,
				Username:  posterUsername,
				AvatarUrl: posterAvatarUrl,
			}
		})

		return &vo.ArticlesSimpleVo{
			Id:             t.Id,
			Title:          t.Title,
			Description:    t.Description,
			LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
			CreateTime:     t.CreatedAt.Format(time.DateTime),
			AuthorId:       t.UserId,
			Username:       username,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			Categories:     categoryNames,
			CategoriesId:   t.CategoryId,
			Type:           t.Type,
			TypeStr:        articlesTypeMap[int(t.Type)].Name,
			Posters:        postersVo,
		}
	})
}
