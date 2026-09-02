package datamigration

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

func TestBackfillCategoryTopicCounts(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&category.Entity{}, &topics.Entity{}, &topicCategoryIndex.Entity{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := conn.Create(&[]category.Entity{{Id: 11, Name: "one", TopicCount: 99}, {Id: 12, Name: "two", TopicCount: 99}}).Error; err != nil {
		t.Fatalf("create categories: %v", err)
	}
	now := time.Now()
	topicRows := []topics.Entity{
		{Id: 21, Title: "published", Status: 1, ProcessStatus: 0},
		{Id: 22, Title: "draft", Status: 0, ProcessStatus: 0},
		{Id: 23, Title: "blocked", Status: 1, ProcessStatus: 1},
		{Id: 24, Title: "deleted", Status: 1, ProcessStatus: 0, DeletedAt: gorm.DeletedAt{Time: now, Valid: true}},
	}
	if err := conn.Unscoped().Create(&topicRows).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}
	indexes := []topicCategoryIndex.Entity{
		{TopicId: 21, CategoryId: 11, Effective: 1},
		{TopicId: 21, CategoryId: 12, Effective: 1},
		{TopicId: 22, CategoryId: 11, Effective: 1},
		{TopicId: 23, CategoryId: 11, Effective: 1},
		{TopicId: 24, CategoryId: 11, Effective: 1},
		{TopicId: 22, CategoryId: 12, Effective: 0},
	}
	if err := conn.Create(&indexes).Error; err != nil {
		t.Fatalf("create category indexes: %v", err)
	}

	result := BackfillCategoryTopicCountsWithDB(conn)
	if result.Failed != 0 || result.Categories != 2 {
		t.Fatalf("BackfillCategoryTopicCountsWithDB() = %#v", result)
	}
	var categories []category.Entity
	if err := conn.Order("id ASC").Find(&categories).Error; err != nil {
		t.Fatalf("load categories: %v", err)
	}
	if categories[0].TopicCount != 1 || categories[1].TopicCount != 1 {
		t.Fatalf("topic counts = %d, %d; want 1, 1", categories[0].TopicCount, categories[1].TopicCount)
	}
}
