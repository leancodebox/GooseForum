package hotdataserve

import (
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/datacache"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"time"
)

// 初始化缓存
var (
	siteStatisticsDataCache = &datacache.Cache[vo.SiteStats]{}
	articleCache            = &datacache.Cache[[]articles.SmallEntity]{}
	articleSimpleDtoCache   = &datacache.Cache[[]vo.ArticlesSimpleDto]{}
	articleCategoryCache    = &datacache.Cache[[]*articleCategory.Entity]{}
	articleCategoryMapCache = &datacache.Cache[map[uint64]*articleCategory.Entity]{}
)

var articlesType = []datastruct.Option[string, int]{
	{Name: "分享", Value: 1},
	{Name: "求助", Value: 2},
}

var articlesTypeMap = array.Slice2Map(articlesType, func(v datastruct.Option[string, int]) int {
	return v.Value
})

func GetArticlesType() *[]datastruct.Option[string, int] {
	return &articlesType
}

func GetArticlesTypeName(iType int) string {
	return articlesTypeMap[iType].Name
}

func GetLatestArticleSimpleDto() []vo.ArticlesSimpleDto {
	return articleSimpleDtoCache.GetOrLoad("home:GetLatestArticles", func() ([]vo.ArticlesSimpleDto, error) {
		return ArticlesSmallEntity2Dto(GetLatestArticles()), nil
	}, time.Second*10)
}

func GetSiteStatisticsData() vo.SiteStats {
	data, _ := siteStatisticsDataCache.GetOrLoadE("", func() (vo.SiteStats, error) {
		configEntity := pageConfig.GetByPageType(pageConfig.FriendShipLinks)
		res := jsonopt.Decode[[]pageConfig.FriendLinksGroup](configEntity.Config)
		linksCount := 0
		for _, group := range res {
			linksCount += len(group.Links)
		}
		return vo.SiteStats{
			UserCount:         users.GetCount(),
			UserMonthCount:    users.GetMonthCount(),
			ArticleCount:      articles.GetCount(),
			ArticleMonthCount: articles.GetMonthCount(),
			Reply:             reply.GetCount(),
			LinksCount:        linksCount,
		}, nil
	}, time.Second*5)
	return data
}

func GetRecommendedArticles() []articles.SmallEntity {
	data, _ := articleCache.GetOrLoadE(
		"GetRecommendedArticles",
		func() ([]articles.SmallEntity, error) {
			return articles.GetRecommendedArticles(4)
		},
		5*time.Minute, // 缓存5分钟
	)
	return data
}

func GetLatestArticles() []articles.SmallEntity {
	data, _ := articleCache.GetOrLoadE(
		"GetLatestArticles",
		func() ([]articles.SmallEntity, error) {
			return articles.GetLatestArticles(20)

		},
		10*time.Second, // 缓存5s
	)
	return data
}

func GetArticleCategory() []*articleCategory.Entity {
	data, _ := articleCategoryCache.GetOrLoadE(
		"GetArticleCategory",
		func() ([]*articleCategory.Entity, error) {
			return articleCategory.All(), nil

		},
		1*time.Minute, // 缓存5分钟
	)
	return data
}

func ArticleCategoryLabel() []datastruct.Option[string, uint64] {
	return array.Map(GetArticleCategory(), func(t *articleCategory.Entity) datastruct.Option[string, uint64] {
		return datastruct.Option[string, uint64]{
			Name:  t.Category,
			Value: t.Id,
		}
	})
}

// GetMapByIds 根据ID列表获取分类Map
func ArticleCategoryMap() map[uint64]*articleCategory.Entity {
	data, _ := articleCategoryMapCache.GetOrLoadE(
		"GetArticleCategory",
		func() (map[uint64]*articleCategory.Entity, error) {
			return array.Slice2Map(articleCategory.All(), func(v *articleCategory.Entity) uint64 {
				return v.Id
			}), nil
		},
		1*time.Minute, // 缓存5分钟
	)
	return data
}

func ArticlesSmallEntity2Dto(data []articles.SmallEntity) []vo.ArticlesSimpleDto {
	userIds := array.Map(data, func(t articles.SmallEntity) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)
	return ArticlesSmallEntityWithUser2Dto(data, userMap)
}
func ArticlesSmallEntityWithUser2Dto(data []articles.SmallEntity, userMap map[uint64]*users.EntityComplete) []vo.ArticlesSimpleDto {
	categoryMap := ArticleCategoryMap()
	return array.Map(data, func(t articles.SmallEntity) vo.ArticlesSimpleDto {
		categoryNames := array.Map(t.CategoryId, func(item uint64) string {
			if category, ok := categoryMap[item]; ok {
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
		return vo.ArticlesSimpleDto{
			Id:             t.Id,
			Title:          t.Title,
			LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
			CreateTime:     t.CreatedAt.Format(time.DateTime),
			AuthorId:       t.UserId,
			Username:       username,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			Category:       array.FirstOr(categoryNames, "未分类"),
			Categories:     categoryNames,
			CategoriesId:   t.CategoryId,
			Type:           t.Type,
			TypeStr:        articlesTypeMap[int(t.Type)].Name,
		}
	})
}
