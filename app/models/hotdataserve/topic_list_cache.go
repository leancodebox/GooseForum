package hotdataserve

import (
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
)

const (
	maxCachedTopicPage = 50
	topicListCacheTTL  = 5 * time.Second
	topicListEntries   = 512
)

type TopicSimpleVoPage struct {
	Topics  []*vo.TopicsSimpleVo
	HasNext bool
}

var topicSimpleVoCache = &localcache.Cache[TopicSimpleVoPage]{MaxEntries: topicListEntries}

func GetLatestTopicsSimpleVoPaginated(page int, sort string) TopicSimpleVoPage {
	page = normalizeTopicPage(page)
	sort = normalizeTopicSort(sort)
	if !shouldCacheTopicPage(page) {
		return loadLatestTopicsSimpleVoPaginated(page, sort)
	}
	key := "home:GetLatestTopics:" + sort + ":" + strconv.Itoa(page)
	return topicSimpleVoCache.GetOrLoad(key, func() (TopicSimpleVoPage, error) {
		return loadLatestTopicsSimpleVoPaginated(page, sort), nil
	}, topicListCacheTTL)
}

func GetTopicsByCategorySimpleVo(categoryId uint64, sort string, page int) TopicSimpleVoPage {
	page = normalizeTopicPage(page)
	sort = normalizeTopicSort(sort)
	if !shouldCacheTopicPage(page) {
		return loadTopicsByCategorySimpleVo(categoryId, sort, page)
	}
	key := "GetTopicsByCategory:" + strconv.FormatUint(categoryId, 10) + ":" + sort + ":" + strconv.Itoa(page)
	return topicSimpleVoCache.GetOrLoad(key, func() (TopicSimpleVoPage, error) {
		return loadTopicsByCategorySimpleVo(categoryId, sort, page), nil
	}, topicListCacheTTL)
}

func normalizeTopicPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeTopicSort(sort string) string {
	switch sort {
	case "hot", "popular", "new":
		return sort
	default:
		return "latest"
	}
}

func shouldCacheTopicPage(page int) bool {
	return page <= maxCachedTopicPage
}

func loadLatestTopicsSimpleVoPaginated(page int, sort string) TopicSimpleVoPage {
	res := topics.Page[topics.SmallEntity](topics.PageQuery{
		Page:         page,
		PageSize:     20,
		FilterStatus: true,
		Sort:         sort,
	})
	return TopicSimpleVoPage{
		Topics:  TopicsSmallEntity2Vo(topicSmallEntitiesToPointers(res.Data)),
		HasNext: res.HasNext,
	}
}

func loadTopicsByCategorySimpleVo(categoryId uint64, sort string, page int) TopicSimpleVoPage {
	res := topics.Page[topics.SmallEntity](topics.PageQuery{
		Page:         page,
		PageSize:     20,
		CategoryId:   categoryId,
		FilterStatus: true,
		Sort:         sort,
	})
	return TopicSimpleVoPage{
		Topics:  TopicsSmallEntity2Vo(topicSmallEntitiesToPointers(res.Data)),
		HasNext: res.HasNext,
	}
}

func topicSmallEntitiesToPointers(data []topics.SmallEntity) []*topics.SmallEntity {
	res := make([]*topics.SmallEntity, 0, len(data))
	for i := range data {
		res = append(res, &data[i])
	}
	return res
}

func TopicsSmallEntity2Vo(data []*topics.SmallEntity) []*vo.TopicsSimpleVo {
	userIDs := make([]uint64, 0, len(data)*2)
	seenUserIDs := make(map[uint64]struct{}, len(data)*2)
	for _, topic := range data {
		if topic == nil {
			continue
		}
		if _, ok := seenUserIDs[topic.UserId]; !ok {
			seenUserIDs[topic.UserId] = struct{}{}
			userIDs = append(userIDs, topic.UserId)
		}
		for _, poster := range topic.GetPosters() {
			if _, ok := seenUserIDs[poster.UserID]; ok {
				continue
			}
			seenUserIDs[poster.UserID] = struct{}{}
			userIDs = append(userIDs, poster.UserID)
		}
	}
	userMap := users.GetMapByIds(userIDs)
	return TopicsSmallEntityWithUser2Vo(data, userMap)
}

func TopicsSmallEntityWithUser2Vo(data []*topics.SmallEntity, userMap map[uint64]*users.EntityComplete) []*vo.TopicsSimpleVo {
	categoryMap := CategoryMap()
	res := make([]*vo.TopicsSimpleVo, 0, len(data))
	for _, t := range data {
		if t == nil {
			continue
		}

		categoryNames := make([]string, 0, len(t.CategoryIds))
		for _, item := range t.CategoryIds {
			if category, ok := categoryMap[item]; ok && category != nil {
				categoryNames = append(categoryNames, category.Name)
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

		res = append(res, &vo.TopicsSimpleVo{
			Id:             t.Id,
			Title:          t.Title,
			Description:    t.Excerpt,
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
			CategoriesId:   t.CategoryIds,
			Type:           0,
			TypeStr:        "",
			ProcessStatus:  t.ProcessStatus,
			Posters:        postersVo,
		})
	}
	return res
}

func ClearTopicListCache() {
	topicSimpleVoCache.Clear()
}
