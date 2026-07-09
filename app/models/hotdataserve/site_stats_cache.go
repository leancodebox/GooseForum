package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/samber/lo"
)

const siteStatsCacheTTL = 5 * time.Second

var siteStatisticsDataCache = &localcache.Cache[*vo.SiteStats]{MaxEntries: finiteCacheEntries}

func GetSiteStatisticsData() *vo.SiteStats {
	return siteStatisticsDataCache.GetOrLoad("", func() (*vo.SiteStats, error) {
		res := GetFriendLinksConfigCache()
		linksCount := lo.SumBy(res, func(group pageConfig.FriendLinksGroup) int {
			return len(group.Links)
		})
		return &vo.SiteStats{
			UserCount:       users.GetMaxId(),
			UserMonthCount:  dailyStats.GetCurrentMonthSum(dailyStats.StatTypeRegCount),
			TopicMaxID:      topics.GetMaxId(),
			TopicMonthCount: dailyStats.GetCurrentMonthSum(dailyStats.StatTypeTopicCount),
			PostMaxID:       posts.GetMaxId(),
			LinksCount:      linksCount,
		}, nil
	}, siteStatsCacheTTL)
}
