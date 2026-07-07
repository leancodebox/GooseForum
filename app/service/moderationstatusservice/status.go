package moderationstatusservice

import (
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/models/forum/reports"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/moderatorservice"
)

const statusTTL = 5 * time.Minute

var statusCache = localcache.Cache[bool]{MaxEntries: 2048}

func HasOpenReports(userID uint64) bool {
	if userID == 0 {
		return false
	}
	if !moderatorservice.CanAccessModeration(userID) {
		return false
	}
	global, categoryIDs := moderatorservice.ScopeForUser(userID)
	if global {
		categoryIDs = allCategoryIDs()
	}
	for _, categoryID := range categoryIDs {
		if hasOpenReport(cacheKeyCategory(categoryID), []uint64{categoryID}) {
			return true
		}
	}
	return false
}

func InvalidateArticle(articleID uint64) {
	InvalidateTopic(articleID)
}

func InvalidateTopic(topicID uint64) {
	topic := topics.GetSimple(topicID)
	for _, categoryID := range topic.CategoryIds {
		statusCache.Delete(cacheKeyCategory(categoryID))
	}
}

func hasOpenReport(key string, categoryIDs []uint64) bool {
	return statusCache.GetOrLoad(key, func() (bool, error) {
		rows := reports.CursorPage(reports.CursorPageQuery{
			Status:           reports.StatusOpen,
			ScopeCategoryIDs: categoryIDs,
			PageSize:         1,
		})
		return len(rows) > 0, nil
	}, statusTTL)
}

func cacheKeyCategory(categoryID uint64) string {
	return "moderation:reports:category:" + strconv.FormatUint(categoryID, 10)
}

func allCategoryIDs() []uint64 {
	categories := hotdataserve.GetCategory()
	ids := make([]uint64, 0, len(categories))
	for _, item := range categories {
		if item != nil && item.Id > 0 {
			ids = append(ids, item.Id)
		}
	}
	return ids
}
