package unreadservice

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestAudienceStatusCacheSeparatesReadableCategorySets(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &eventNotification.Entity{}); err != nil {
		t.Fatalf("migrate unread audience tables: %v", err)
	}
	const userID uint64 = 976001
	const topicID uint64 = 976002
	const readableCategoryID uint64 = 976003
	const otherCategoryID uint64 = 976004
	conn.Where("user_id = ?", userID).Delete(&eventNotification.Entity{})
	conn.Unscoped().Delete(&topics.Entity{}, topicID)
	statusCache.Clear()
	if err := conn.Create(&topics.Entity{Id: topicID, Title: "Audience notification", Status: 1, MainCategoryId: readableCategoryID, CategoryIds: []uint64{readableCategoryID}}).Error; err != nil {
		t.Fatalf("create notification topic: %v", err)
	}
	if err := conn.Create(&eventNotification.Entity{UserId: userID, TopicId: topicID, EventType: eventNotification.EventTypeComment}).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	if got := GetStatusForAudience(userID, []uint64{otherCategoryID}, true); got.Notifications {
		t.Fatalf("unread status for unrelated category = %+v", got)
	}
	if got := GetStatusForAudience(userID, []uint64{readableCategoryID}, true); !got.Notifications {
		t.Fatalf("unread status for readable category = %+v", got)
	}
	if left, right := audienceCacheKey(userID, []uint64{2, 1, 2}, true), audienceCacheKey(userID, []uint64{1, 2}, true); left != right {
		t.Fatalf("equivalent audience keys differ: %q != %q", left, right)
	}
	before := audienceCacheKey(userID, []uint64{readableCategoryID}, true)
	Invalidate(userID)
	if after := audienceCacheKey(userID, []uint64{readableCategoryID}, true); after == before {
		t.Fatalf("audience generation did not change after invalidation: %q", after)
	}
}
