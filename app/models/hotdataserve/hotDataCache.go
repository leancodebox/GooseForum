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
		return ArticlesSmallEntity2Vo(smallEntitiesToPointers(res.Data)), nil
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
		return ArticlesSmallEntity2Vo(smallEntitiesToPointers(res.Data)), nil
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
			return smallEntitiesToPointers(res), err
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
			return smallEntitiesToPointers(res), err
		},
		10*time.Second,
	)
	return data
}

func smallEntitiesToPointers(data []articles.SmallEntity) []*articles.SmallEntity {
	res := make([]*articles.SmallEntity, 0, len(data))
	for i := range data {
		res = append(res, &data[i])
	}
	return res
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
	userIDs := make([]uint64, 0, len(data)*2)
	seenUserIDs := make(map[uint64]struct{}, len(data)*2)
	for _, article := range data {
		if article == nil {
			continue
		}
		if _, ok := seenUserIDs[article.UserId]; !ok {
			seenUserIDs[article.UserId] = struct{}{}
			userIDs = append(userIDs, article.UserId)
		}
		for _, poster := range article.GetPosters() {
			if _, ok := seenUserIDs[poster.UserID]; ok {
				continue
			}
			seenUserIDs[poster.UserID] = struct{}{}
			userIDs = append(userIDs, poster.UserID)
		}
	}
	userMap := users.GetMapByIds(userIDs)
	return ArticlesSmallEntityWithUser2Vo(data, userMap)
}

func ArticlesSmallEntityWithUser2Vo(data []*articles.SmallEntity, userMap map[uint64]*users.EntityComplete) []*vo.ArticlesSimpleVo {
	categoryMap := ArticleCategoryMap()
	res := make([]*vo.ArticlesSimpleVo, 0, len(data))
	for _, t := range data {
		if t == nil {
			continue
		}

		categoryNames := make([]string, 0, len(t.CategoryId))
		for _, item := range t.CategoryId {
			if category, ok := categoryMap[item]; ok && category != nil {
				categoryNames = append(categoryNames, category.Category)
				continue
			}
			categoryNames = append(categoryNames, "")
		}

		username := ""
		avatarUrl := urlconfig.GetDefaultAvatar()
		if user, ok := userMap[t.UserId]; ok {
			username = user.Username
			avatarUrl = user.GetWebAvatarUrl()
		}

		posters := t.GetPosters()
		postersVo := make([]vo.PosterVo, 0, len(posters))
		for _, poster := range posters {
			posterUsername := ""
			posterAvatarUrl := urlconfig.GetDefaultAvatar()
			if user, ok := userMap[poster.UserID]; ok {
				posterUsername = user.Username
				posterAvatarUrl = user.GetWebAvatarUrl()
			}
			postersVo = append(postersVo, vo.PosterVo{
				Id:        poster.UserID,
				Username:  posterUsername,
				AvatarUrl: posterAvatarUrl,
			})
		}

		res = append(res, &vo.ArticlesSimpleVo{
			Id:             t.Id,
			Title:          t.Title,
			Description:    t.Description,
			FirstImageURL:  t.FirstImageURL,
			LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
			CreateTime:     t.CreatedAt.Format(time.DateTime),
			AuthorId:       t.UserId,
			Username:       username,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			PinWeight:      t.PinWeight,
			Categories:     categoryNames,
			CategoriesId:   t.CategoryId,
			Type:           t.Type,
			TypeStr:        articlesTypeMap[int(t.Type)].Name,
			Posters:        postersVo,
		})
	}
	return res
}

func ClearArticleListCache() {
	articleCache.Clear()
	articleSimpleVoCache.Clear()
}
