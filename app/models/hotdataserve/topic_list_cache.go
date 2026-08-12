package hotdataserve

import (
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

const (
	maxCachedTopicPage = 32
	topicListCacheTTL  = 5 * time.Second
)

type TopicSimpleVoPage struct {
	Topics  []*vo.TopicsSimpleVo
	HasNext bool
}

var topicSimpleVoCache = &localcache.Cache[TopicSimpleVoPage]{MaxEntries: cacheconfig.Current().TopicList}

func GetLatestTopicsSimpleVoPaginated(page int, sort string) TopicSimpleVoPage {
	return GetLatestTopicsSimpleVoPaginatedForAudience(page, sort, "public", nil, false, true)
}

func GetLatestTopicsSimpleVoPaginatedForAudience(
	page int,
	sort string,
	audience string,
	readableCategoryIDs []uint64,
	filterAudience bool,
	cacheable bool,
) TopicSimpleVoPage {
	page = normalizeTopicPage(page)
	sort = normalizeTopicSort(sort)
	if !cacheable || audience == "" || !shouldCacheTopicPage(page) {
		return loadLatestTopicsSimpleVoPaginated(page, sort, readableCategoryIDs, filterAudience)
	}
	key := "home:GetLatestTopics:" + audience + ":" + sort + ":" + strconv.Itoa(page)
	return topicSimpleVoCache.GetOrLoad(key, func() (TopicSimpleVoPage, error) {
		return loadLatestTopicsSimpleVoPaginated(page, sort, readableCategoryIDs, filterAudience), nil
	}, topicListCacheTTL)
}

// GetTopicsByCategorySimpleVo lists a category page. The category being browsed
// may be an auxiliary tag on topics whose main category the viewer cannot read,
// so the audience filter still applies here — it is the one place where the
// tag being listed and the category deciding visibility differ.
func GetTopicsByCategorySimpleVo(
	categoryId uint64,
	sort string,
	page int,
	audience string,
	readableCategoryIDs []uint64,
	filterAudience bool,
	cacheable bool,
) TopicSimpleVoPage {
	page = normalizeTopicPage(page)
	sort = normalizeTopicSort(sort)
	if !cacheable || audience == "" || !shouldCacheTopicPage(page) {
		return loadTopicsByCategorySimpleVo(categoryId, sort, page, readableCategoryIDs, filterAudience)
	}
	key := "GetTopicsByCategory:" + audience + ":" + strconv.FormatUint(categoryId, 10) + ":" + sort + ":" + strconv.Itoa(page)
	return topicSimpleVoCache.GetOrLoad(key, func() (TopicSimpleVoPage, error) {
		return loadTopicsByCategorySimpleVo(categoryId, sort, page, readableCategoryIDs, filterAudience), nil
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

func loadLatestTopicsSimpleVoPaginated(page int, sort string, readableCategoryIDs []uint64, filterAudience bool) TopicSimpleVoPage {
	res := topics.Page(topics.PageQuery{
		Page:                page,
		PageSize:            20,
		FilterStatus:        true,
		FilterAudience:      filterAudience,
		ReadableCategoryIds: readableCategoryIDs,
		Sort:                sort,
	})
	return TopicSimpleVoPage{
		Topics:  transform.Topics2Vo(topicEntitiesToPointers(res.Data), CategoryMap()),
		HasNext: res.HasNext,
	}
}

func loadTopicsByCategorySimpleVo(categoryId uint64, sort string, page int, readableCategoryIDs []uint64, filterAudience bool) TopicSimpleVoPage {
	res := topics.Page(topics.PageQuery{
		Page:                page,
		PageSize:            20,
		CategoryId:          categoryId,
		FilterStatus:        true,
		FilterAudience:      filterAudience,
		ReadableCategoryIds: readableCategoryIDs,
		Sort:                sort,
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
