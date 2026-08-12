package forum

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

// The point of reading visibility from a single column is that the list filter
// and the detail check are the same predicate, so they cannot disagree — no
// title that is visible in a list can 404 when clicked, and none that is hidden
// from a list can be opened. This test asserts that agreement in both
// directions for a topic that spans a public and a restricted category.
func TestListAndDetailAgreeOnMainCategory(t *testing.T) {
	const (
		publicCategoryID      = uint64(994301)
		restrictedCategoryID  = uint64(994302)
		publicMainTopicID     = uint64(994310)
		restrictedMainTopicID = uint64(994311)
		authorID              = uint64(994320)
	)

	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &topicCategoryIndex.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatalf("migrate main category tables: %v", err)
	}
	ensureForumTestAccessCategory(t, publicCategoryID)
	ensureForumTestAccessCategory(t, restrictedCategoryID)
	revokeForumTestEveryoneRead(t, restrictedCategoryID)

	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	topicIDs := []uint64{publicMainTopicID, restrictedMainTopicID}
	cleanup := func() {
		conn.Unscoped().Delete(&users.EntityComplete{}, authorID)
		conn.Unscoped().Delete(&topics.Entity{}, topicIDs)
		conn.Where("topic_id in ?", topicIDs).Delete(&topicCategoryIndex.Entity{})
		hotdataserve.ClearTopicListCache()
		hotdataserve.ClearTopicCategoryCache()
	}
	cleanup()
	t.Cleanup(cleanup)

	conn.Create(&users.EntityComplete{Id: authorID, Username: "main-category-author"})
	// Both topics carry both categories. Only the order differs.
	conn.Create(&topics.Entity{
		Id: publicMainTopicID, Title: "public main", UserId: authorID, Status: 1, ProcessStatus: 0,
		CategoryIds:    []uint64{publicCategoryID, restrictedCategoryID},
		MainCategoryId: publicCategoryID,
		CreatedAt:      now, UpdatedAt: now,
	})
	conn.Create(&topics.Entity{
		Id: restrictedMainTopicID, Title: "restricted main", UserId: authorID, Status: 1, ProcessStatus: 0,
		CategoryIds:    []uint64{restrictedCategoryID, publicCategoryID},
		MainCategoryId: restrictedCategoryID,
		CreatedAt:      now, UpdatedAt: now,
	})
	for _, topicID := range topicIDs {
		conn.Create(&topicCategoryIndex.Entity{TopicId: topicID, CategoryId: publicCategoryID, Effective: 1})
		conn.Create(&topicCategoryIndex.Entity{TopicId: topicID, CategoryId: restrictedCategoryID, Effective: 1})
	}
	hotdataserve.ClearTopicListCache()
	hotdataserve.ClearTopicCategoryCache()

	guest, err := accesscontrol.Resolve(0)
	if err != nil {
		t.Fatalf("resolve guest snapshot: %v", err)
	}
	readable := guest.ReadableCategoryIDs()

	listed := topics.Page(topics.PageQuery{
		Page: 1, PageSize: 100, FilterStatus: true,
		FilterAudience: true, ReadableCategoryIds: readable, Sort: "new",
	})
	inList := map[uint64]bool{}
	for _, item := range listed.Data {
		inList[item.Id] = true
	}

	for _, tc := range []struct {
		topicID uint64
		want    bool
		name    string
	}{
		{publicMainTopicID, true, "public main category"},
		{restrictedMainTopicID, false, "restricted main category"},
	} {
		topic := topics.GetSimple(tc.topicID)
		canOpen := guest.CanReadCategory(topic.MainCategoryId)
		if canOpen != tc.want {
			t.Fatalf("%s: detail readable = %v, want %v", tc.name, canOpen, tc.want)
		}
		if inList[tc.topicID] != tc.want {
			t.Fatalf("%s: listed = %v, want %v", tc.name, inList[tc.topicID], tc.want)
		}
		if inList[tc.topicID] != canOpen {
			t.Fatalf("%s: list and detail disagree", tc.name)
		}
	}

	// The one rule the model still needs: the public category is an auxiliary
	// tag on the restricted-main topic, so browsing it must not surface that
	// topic even though the tag itself is readable.
	categoryPage := hotdataserve.GetTopicsByCategorySimpleVo(publicCategoryID, "new", 1, "public", readable, true, false)
	for _, item := range categoryPage.Topics {
		if item != nil && item.Id == restrictedMainTopicID {
			t.Fatal("auxiliary category listing leaked a topic whose main category is restricted")
		}
	}
}
