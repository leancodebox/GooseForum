package hotdataserve

import (
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
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
	res := topics.Page(topics.PageQuery{
		Page:         page,
		PageSize:     20,
		FilterStatus: true,
		Sort:         sort,
	})
	return TopicSimpleVoPage{
		Topics:  transform.Topics2Vo(topicEntitiesToPointers(res.Data), CategoryMap()),
		HasNext: res.HasNext,
	}
}

func loadTopicsByCategorySimpleVo(categoryId uint64, sort string, page int) TopicSimpleVoPage {
	res := topics.Page(topics.PageQuery{
		Page:         page,
		PageSize:     20,
		CategoryId:   categoryId,
		FilterStatus: true,
		Sort:         sort,
	})
	return TopicSimpleVoPage{
		Topics:  transform.Topics2Vo(topicEntitiesToPointers(res.Data), CategoryMap()),
		HasNext: res.HasNext,
	}
}

func topicEntitiesToPointers(data []topics.Entity) []*topics.Entity {
	res := make([]*topics.Entity, 0, len(data))
	for i := range data {
		res = append(res, &data[i])
	}
	return res
}

func ClearTopicListCache() {
	topicSimpleVoCache.Clear()
}
