package topicservice

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"gorm.io/gorm"
)

func TestSaveTopicAndFirstPostIsAtomic(t *testing.T) {
	conn := openTopicWriteDB(t)
	if err := conn.Create(&posts.Entity{Id: 77, TopicId: 999, PostNo: 1}).Error; err != nil {
		t.Fatalf("create conflicting post: %v", err)
	}
	topic := &topics.Entity{Id: 10, Title: "atomic", UserId: 1, Status: 1}
	firstPost := &posts.Entity{Id: 77, UserId: 1, Content: "body"}
	err := SaveTopicAndFirstPostWithDB(conn, FirstPostWrite{
		Topic:       topic,
		FirstPost:   firstPost,
		CategoryIDs: []uint64{3},
		Create:      true,
	})
	if err == nil {
		t.Fatal("conflicting post write error = nil")
	}
	var count int64
	if err := conn.Model(&topics.Entity{}).Where("id = ?", 10).Count(&count).Error; err != nil {
		t.Fatalf("count rolled back topic: %v", err)
	}
	if count != 0 {
		t.Fatalf("topic count after rollback = %d", count)
	}
}

func TestSaveTopicAndFirstPostPersistsBothCategoryRepresentations(t *testing.T) {
	conn := openTopicWriteDB(t)
	topic := &topics.Entity{Title: "consistent", UserId: 1, Status: 1, Posters: []topics.Poster{{UserID: 1}}}
	firstPost := &posts.Entity{UserId: 1, Content: "body"}
	if err := SaveTopicAndFirstPostWithDB(conn, FirstPostWrite{
		Topic:       topic,
		FirstPost:   firstPost,
		CategoryIDs: []uint64{3, 4},
		Create:      true,
	}); err != nil {
		t.Fatalf("SaveTopicAndFirstPostWithDB: %v", err)
	}
	var stored topics.Entity
	if err := conn.First(&stored, topic.Id).Error; err != nil {
		t.Fatalf("load topic: %v", err)
	}
	if len(stored.CategoryIds) != 2 || stored.CategoryIds[0] != 3 || stored.CategoryIds[1] != 4 {
		t.Fatalf("stored category ids = %v", stored.CategoryIds)
	}
	var indexes []topicCategoryIndex.Entity
	if err := conn.Where("topic_id = ? AND effective = 1", topic.Id).Order("category_id ASC").Find(&indexes).Error; err != nil {
		t.Fatalf("load category indexes: %v", err)
	}
	if len(indexes) != 2 || indexes[0].CategoryId != 3 || indexes[1].CategoryId != 4 {
		t.Fatalf("stored category indexes = %#v", indexes)
	}
}

func TestSaveTopicAndFirstPostRechecksRestrictedSelectionInTransaction(t *testing.T) {
	conn := openTopicWriteDB(t)
	topic := &topics.Entity{Title: "restricted mix", UserId: 1, Status: 1}
	firstPost := &posts.Entity{UserId: 1, Content: "body"}
	err := SaveTopicAndFirstPostWithDB(conn, FirstPostWrite{
		Topic: topic, FirstPost: firstPost, CategoryIDs: []uint64{3, 5}, Create: true,
	})
	if !errors.Is(err, accesscontrol.ErrRestrictedCategorySingle) {
		t.Fatalf("restricted mixed write error = %v", err)
	}
	var topicCount, postCount, indexCount int64
	conn.Model(&topics.Entity{}).Count(&topicCount)
	conn.Model(&posts.Entity{}).Count(&postCount)
	conn.Model(&topicCategoryIndex.Entity{}).Count(&indexCount)
	if topicCount != 0 || postCount != 0 || indexCount != 0 {
		t.Fatalf("rejected write persisted topic=%d post=%d indexes=%d", topicCount, postCount, indexCount)
	}
}

func TestSaveTopicAndFirstPostCanonicalizesAndBoundsCategories(t *testing.T) {
	conn := openTopicWriteDB(t)
	topic := &topics.Entity{Title: "canonical", UserId: 1, Status: 1}
	firstPost := &posts.Entity{UserId: 1, Content: "body"}
	if err := SaveTopicAndFirstPostWithDB(conn, FirstPostWrite{
		Topic: topic, FirstPost: firstPost, CategoryIDs: []uint64{0, 3, 3, 4}, Create: true,
	}); err != nil {
		t.Fatalf("canonical category write: %v", err)
	}
	if len(topic.CategoryIds) != 2 || topic.CategoryIds[0] != 3 || topic.CategoryIds[1] != 4 || topic.MainCategoryId != 3 {
		t.Fatalf("canonical categories = %v main=%d", topic.CategoryIds, topic.MainCategoryId)
	}

	err := SaveTopicAndFirstPostWithDB(conn, FirstPostWrite{
		Topic: &topics.Entity{Title: "too many"}, FirstPost: &posts.Entity{Content: "body"},
		CategoryIDs: []uint64{1, 2, 3, 4}, Create: true,
	})
	if !errors.Is(err, accesscontrol.ErrTooManyCategories) {
		t.Fatalf("too many categories error = %v", err)
	}
}

func openTopicWriteDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(
		&accessGroups.Entity{}, &categoryGroupPermissions.Entity{},
		&topics.Entity{}, &posts.Entity{}, &topicCategoryIndex.Entity{},
	); err != nil {
		t.Fatalf("migrate topic write tables: %v", err)
	}
	everyoneKey := accessGroups.SystemKeyEveryone
	everyone := accessGroups.Entity{Id: 1, Name: "Everyone", SystemKey: &everyoneKey, JoinMode: accessGroups.JoinModeSystem, Status: accessGroups.StatusEnabled}
	if err := conn.Create(&everyone).Error; err != nil {
		t.Fatalf("create everyone group: %v", err)
	}
	for _, categoryID := range []uint64{3, 4} {
		grant := categoryGroupPermissions.Entity{
			CategoryId: categoryID, AccessGroupId: everyone.Id,
			PermissionLevel: categoryGroupPermissions.PermissionRead,
			Status:          categoryGroupPermissions.StatusEnabled,
		}
		if err := conn.Create(&grant).Error; err != nil {
			t.Fatalf("create public category grant: %v", err)
		}
	}
	return conn
}
