package hotdataserve

import (
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
)

const (
	maxCachedArticlePage = 50
	articleListCacheTTL  = 5 * time.Second
	articleListEntries   = 512
)

type ArticleSimpleVoPage struct {
	Topics  []*vo.ArticlesSimpleVo
	HasNext bool
}

var articleSimpleVoCache = &localcache.Cache[ArticleSimpleVoPage]{MaxEntries: articleListEntries}

func GetLatestArticlesSimpleVoPaginated(page int, sort string) ArticleSimpleVoPage {
	page = normalizeArticlePage(page)
	sort = normalizeArticleSort(sort)
	if !shouldCacheArticlePage(page) {
		return loadLatestArticlesSimpleVoPaginated(page, sort)
	}
	key := "home:GetLatestArticles:" + sort + ":" + strconv.Itoa(page)
	return articleSimpleVoCache.GetOrLoad(key, func() (ArticleSimpleVoPage, error) {
		return loadLatestArticlesSimpleVoPaginated(page, sort), nil
	}, articleListCacheTTL)
}

func GetArticlesByCategorySimpleVo(categoryId uint64, sort string, page int) ArticleSimpleVoPage {
	page = normalizeArticlePage(page)
	sort = normalizeArticleSort(sort)
	if !shouldCacheArticlePage(page) {
		return loadArticlesByCategorySimpleVo(categoryId, sort, page)
	}
	key := "GetArticlesByCategory:" + strconv.FormatUint(categoryId, 10) + ":" + sort + ":" + strconv.Itoa(page)
	return articleSimpleVoCache.GetOrLoad(key, func() (ArticleSimpleVoPage, error) {
		return loadArticlesByCategorySimpleVo(categoryId, sort, page), nil
	}, articleListCacheTTL)
}

func normalizeArticlePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeArticleSort(sort string) string {
	switch sort {
	case "hot", "popular", "new":
		return sort
	default:
		return "latest"
	}
}

func shouldCacheArticlePage(page int) bool {
	return page <= maxCachedArticlePage
}

func loadLatestArticlesSimpleVoPaginated(page int, sort string) ArticleSimpleVoPage {
	res := articles.Page[articles.SmallEntity](articles.PageQuery{
		Page:         page,
		PageSize:     20,
		FilterStatus: true,
		Sort:         sort,
	})
	return ArticleSimpleVoPage{
		Topics:  ArticlesSmallEntity2Vo(smallEntitiesToPointers(res.Data)),
		HasNext: res.HasNext,
	}
}

func loadArticlesByCategorySimpleVo(categoryId uint64, sort string, page int) ArticleSimpleVoPage {
	res := articles.Page[articles.SmallEntity](articles.PageQuery{
		Page:         page,
		PageSize:     20,
		CategoryId:   categoryId,
		FilterStatus: true,
		Sort:         sort,
	})
	return ArticleSimpleVoPage{
		Topics:  ArticlesSmallEntity2Vo(smallEntitiesToPointers(res.Data)),
		HasNext: res.HasNext,
	}
}

func smallEntitiesToPointers(data []articles.SmallEntity) []*articles.SmallEntity {
	res := make([]*articles.SmallEntity, 0, len(data))
	for i := range data {
		res = append(res, &data[i])
	}
	return res
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
			ProcessStatus:  t.ProcessStatus,
			Posters:        postersVo,
		})
	}
	return res
}

func ClearArticleListCache() {
	articleSimpleVoCache.Clear()
}
