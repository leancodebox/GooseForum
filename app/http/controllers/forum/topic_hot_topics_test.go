package forum

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

func TestBuildTopicHotTopicsHidesRestrictedTopicsFromGuests(t *testing.T) {
	const (
		publicCategoryID     = uint64(994201)
		restrictedCategoryID = uint64(994202)
		publicTopicID        = uint64(994210)
		restrictedTopicID    = uint64(994211)
		authorID             = uint64(994220)
	)

	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &topicCategoryIndex.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatalf("migrate hot topics tables: %v", err)
	}
	ensureForumTestAccessCategory(t, publicCategoryID)
	ensureForumTestAccessCategory(t, restrictedCategoryID)
	revokeForumTestEveryoneRead(t, restrictedCategoryID)

	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	conn.Unscoped().Delete(&users.EntityComplete{}, authorID)
	conn.Unscoped().Delete(&topics.Entity{}, []uint64{publicTopicID, restrictedTopicID})
	conn.Where("topic_id in ?", []uint64{publicTopicID, restrictedTopicID}).Delete(&topicCategoryIndex.Entity{})
	t.Cleanup(func() {
		conn.Unscoped().Delete(&users.EntityComplete{}, authorID)
		conn.Unscoped().Delete(&topics.Entity{}, []uint64{publicTopicID, restrictedTopicID})
		conn.Where("topic_id in ?", []uint64{publicTopicID, restrictedTopicID}).Delete(&topicCategoryIndex.Entity{})
		hotdataserve.ClearTopicListCache()
		hotdataserve.ClearTopicCategoryCache()
	})

	conn.Create(&users.EntityComplete{Id: authorID, Username: "hot-topics-author"})
	// The reply counts put both topics at the top of the "hot" ordering, restricted first.
	conn.Create(&topics.Entity{
		Id: publicTopicID, Title: "public hot topic", CategoryIds: []uint64{publicCategoryID},
		UserId: authorID, Status: 1, ProcessStatus: 0, ReplyCount: 900001, CreatedAt: now, UpdatedAt: now,
	})
	conn.Create(&topics.Entity{
		Id: restrictedTopicID, Title: "restricted hot topic", CategoryIds: []uint64{restrictedCategoryID},
		UserId: authorID, Status: 1, ProcessStatus: 0, ReplyCount: 900002, CreatedAt: now, UpdatedAt: now,
	})
	conn.Create(&topicCategoryIndex.Entity{TopicId: publicTopicID, CategoryId: publicCategoryID, Effective: 1})
	conn.Create(&topicCategoryIndex.Entity{TopicId: restrictedTopicID, CategoryId: restrictedCategoryID, Effective: 1})
	hotdataserve.ClearTopicListCache()
	hotdataserve.ClearTopicCategoryCache()

	guest, err := accesscontrol.Resolve(0)
	if err != nil {
		t.Fatalf("resolve guest snapshot: %v", err)
	}
	if guest.CanReadCategory(restrictedCategoryID) {
		t.Fatal("test setup: guests must not be able to read the restricted category")
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/t/topic/994210", nil)
	c.Set(accessSnapshotContextKey, guest)

	hotTopics := buildTopicHotTopics(c, 0)

	sawPublic := false
	for _, item := range hotTopics {
		if item.ID == restrictedTopicID {
			t.Fatalf("hot topics leaked restricted topic %q to a guest", item.Title)
		}
		if item.ID == publicTopicID {
			sawPublic = true
		}
	}
	if !sawPublic {
		t.Fatalf("hot topics dropped the readable topic: %#v", hotTopics)
	}
}

// revokeForumTestEveryoneRead turns a category created by ensureForumTestAccessCategory
// into a restricted one by dropping the everyone group's grant on it.
func revokeForumTestEveryoneRead(t *testing.T, categoryID uint64) {
	t.Helper()
	conn := dbconnect.Connect()
	groups, err := accessGroups.GetBySystemKeys([]string{accessGroups.SystemKeyEveryone})
	if err != nil {
		t.Fatalf("load everyone access group: %v", err)
	}
	if len(groups) == 0 {
		t.Fatal("everyone access group missing after backfill")
	}
	for _, group := range groups {
		if err := conn.Where("category_id = ? and access_group_id = ?", categoryID, group.Id).
			Delete(&categoryGroupPermissions.Entity{}).Error; err != nil {
			t.Fatalf("revoke everyone grant: %v", err)
		}
		accesscontrol.InvalidateGroup(group.Id)
	}
	accesscontrol.InvalidateSystemGroups()
}
