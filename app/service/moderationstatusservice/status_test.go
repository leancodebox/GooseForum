package moderationstatusservice

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/reports"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func TestInvalidateTopicClearsCategoryStatusCache(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &category.Entity{}, &topicCategoryIndex.Entity{}, &reports.Entity{}); err != nil {
		t.Fatalf("migrate moderation status tables: %v", err)
	}
	conn.Unscoped().Delete(&topics.Entity{}, 970010)
	conn.Unscoped().Delete(&category.Entity{}, 970003)
	conn.Unscoped().Where("topic_id = ?", 970010).Delete(&topicCategoryIndex.Entity{})
	conn.Where("topic_id = ?", 970010).Delete(&reports.Entity{})
	statusCache.Clear()
	hotdataserve.ClearCategoryCache()

	conn.Create(&category.Entity{Id: 970003, Name: "Moderated", Slug: "moderated"})
	conn.Create(&topics.Entity{Id: 970010, Title: "reported topic", CategoryIds: []uint64{970003}, Status: 1})
	conn.Create(&topicCategoryIndex.Entity{TopicId: 970010, CategoryId: 970003, Effective: 1})

	key := cacheKeyCategory(970003)
	if got := hasOpenReport(key, []uint64{970003}); got {
		t.Fatal("open report cache before report = true, want false")
	}
	conn.Create(&reports.Entity{TargetType: reports.TargetTopic, TargetId: 970010, TopicId: 970010, ReporterId: 1, Reason: reports.ReasonSpam, Status: reports.StatusOpen})
	if got := hasOpenReport(key, []uint64{970003}); got {
		t.Fatal("open report cache before invalidation = true, want cached false")
	}
	InvalidateTopic(970010)
	if got := hasOpenReport(key, []uint64{970003}); !got {
		t.Fatal("open report cache after invalidation = false, want true")
	}
}
