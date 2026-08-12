package api

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/moderationLog"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
	"gorm.io/gorm"
)

func setupAdminTopicTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(
		&accessGroups.Entity{},
		&accessGroupMembers.Entity{},
		&categoryGroupPermissions.Entity{},
		&users.EntityComplete{},
		&topics.Entity{},
		&posts.Entity{},
		&category.Entity{},
		&topicCategoryIndex.Entity{},
		&optRecord.Entity{},
		&moderationLog.Entity{},
	); err != nil {
		t.Fatalf("migrate admin topic tables: %v", err)
	}
	ensureAdminTopicAccess(t, conn)
	return conn
}

func ensureAdminTopicAccess(t *testing.T, conn *gorm.DB) {
	t.Helper()
	result := datamigration.BackfillAccessControlDefaultsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("backfill access defaults: %s", result.LastFailed)
	}
	groups, err := accessGroups.All()
	if err != nil {
		t.Fatalf("load access groups: %v", err)
	}
	accesscontrol.InvalidateSystemGroups()
	for _, group := range groups {
		accesscontrol.InvalidateGroup(group.Id)
	}
}

func seedAdminTopic(t *testing.T, conn *gorm.DB, topicID uint64) (uint64, uint64) {
	t.Helper()
	userID := topicID + 10
	firstPostID := topicID + 100
	categoryID := topicID + 1000
	now := time.Date(2026, 7, 7, 15, 0, 0, 0, time.UTC)
	if err := conn.Create(&users.EntityComplete{Id: userID, Username: "author"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := conn.Create(&category.Entity{Id: categoryID, Name: "General", Slug: "general"}).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	ensureAdminTopicAccess(t, conn)
	topic := topics.Entity{
		Id:             topicID,
		Title:          "Topic title",
		Excerpt:        "Topic excerpt",
		CategoryIds:    []uint64{categoryID},
		MainCategoryId: categoryID,
		UserId:         userID,
		Status:         1,
		ProcessStatus:  0,
		PostCount:      1,
		PostSeq:        1,
		FirstPostId:    firstPostID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := conn.Create(&topic).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if err := conn.Create(&posts.Entity{Id: firstPostID, TopicId: topicID, PostNo: 1, UserId: userID, Content: "first post source", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create first post: %v", err)
	}
	if err := conn.Create(&topicCategoryIndex.Entity{TopicId: topicID, CategoryId: categoryID, Effective: 1}).Error; err != nil {
		t.Fatalf("create topic category index: %v", err)
	}
	return userID, categoryID
}

func TestAdminTopicsListReadsTopics(t *testing.T) {
	conn := setupAdminTopicTestDB(t)
	seedAdminTopic(t, conn, 920001)

	res := TopicsList(component.BetterRequest[TopicsListReq]{Params: TopicsListReq{Page: 1, PageSize: 10}})
	page, ok := res.Data.Result.(component.Page[TopicInfoAdminVo])
	if !ok {
		t.Fatalf("result type = %T", res.Data.Result)
	}
	if len(page.List) == 0 || page.List[0].Id != 920001 || page.List[0].Description != "Topic excerpt" || page.List[0].TopicStatus != 1 {
		t.Fatalf("page = %#v", page)
	}
}

func TestAdminTopicSourceReadsFirstPost(t *testing.T) {
	conn := setupAdminTopicTestDB(t)
	_, categoryID := seedAdminTopic(t, conn, 921001)

	res := TopicSource(component.BetterRequest[TopicSourceReq]{Params: TopicSourceReq{TopicId: 921001}})
	source, ok := res.Data.Result.(TopicSourceVo)
	if !ok {
		t.Fatalf("result type = %T", res.Data.Result)
	}
	if source.Id != 921001 || source.Content != "first post source" || len(source.CategoryId) != 1 || source.CategoryId[0] != categoryID {
		t.Fatalf("source = %#v", source)
	}
}

func TestAdminEditTopicMutatesTopic(t *testing.T) {
	conn := setupAdminTopicTestDB(t)
	_, categoryID := seedAdminTopic(t, conn, 922001)
	if err := conn.Create(&category.Entity{Id: 922999, Name: "Second", Slug: "second"}).Error; err != nil {
		t.Fatalf("create second category: %v", err)
	}
	ensureAdminTopicAccess(t, conn)
	EditTopic(component.BetterRequest[EditTopicReq]{UserId: 1, Params: EditTopicReq{TopicId: 922001, ProcessStatus: 1}})
	topic := topics.Get(922001)
	if topic.ProcessStatus != 1 {
		t.Fatalf("process status = %d, want 1", topic.ProcessStatus)
	}

	EditTopicPin(component.BetterRequest[EditTopicPinReq]{UserId: 1, Params: EditTopicPinReq{TopicId: 922001, PinWeight: 9}})
	topic = topics.Get(922001)
	if topic.PinWeight != 9 {
		t.Fatalf("pin weight = %d, want 9", topic.PinWeight)
	}

	response := EditTopicCategories(component.BetterRequest[EditTopicCategoriesReq]{UserId: 1, Params: EditTopicCategoriesReq{TopicId: 922001, CategoryId: []uint64{922999}}})
	if response.Data.Code != component.SUCCESS {
		t.Fatalf("edit topic categories response = %#v", response.Data)
	}
	topic = topics.Get(922001)
	if len(topic.CategoryIds) != 1 || topic.CategoryIds[0] != 922999 {
		t.Fatalf("topic categories = %#v", topic.CategoryIds)
	}
	indexes := topicCategoryIndex.GetByTopicId(922001)
	active := map[uint64]int{}
	for _, item := range indexes {
		active[item.CategoryId] = item.Effective
	}
	if active[categoryID] != 0 || active[922999] != 1 {
		t.Fatalf("category index active map = %#v", active)
	}
}
