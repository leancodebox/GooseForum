package topicCategoryIndex

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

type categoryCountTestTopic struct {
	Id        uint64 `gorm:"primaryKey;column:id"`
	DeletedAt gorm.DeletedAt
}

func (categoryCountTestTopic) TableName() string { return "topics" }

func TestTopicCategoryIndexRepositoryParity(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate topic category index: %v", err)
	}
	conn.Where("1 = 1").Delete(&Entity{})

	SaveOrCreateById(&Entity{TopicId: 10, CategoryId: 3, Effective: 1})
	SaveOrCreateById(&Entity{TopicId: 20, CategoryId: 3, Effective: 0})
	SaveOrCreateById(&Entity{TopicId: 10, CategoryId: 4, Effective: 1})

	rows := GetByTopicId(10)
	if len(rows) != 2 {
		t.Fatalf("GetByTopicId() len=%d, want 2", len(rows))
	}
	if got := GetOneByCategoryId(3); got.TopicId != 10 {
		t.Fatalf("GetOneByCategoryId(3).TopicId=%d, want 10", got.TopicId)
	}
	if deleted := DeleteByTopicId(10); deleted != 2 {
		t.Fatalf("DeleteByTopicId()=%d, want 2", deleted)
	}
	if rows := GetByTopicId(10); len(rows) != 0 {
		t.Fatalf("GetByTopicId() after delete len=%d, want 0", len(rows))
	}
}

func TestMultiCategoryTopicCountsIgnoresSingleAndDeletedTopics(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&categoryCountTestTopic{}, &Entity{}); err != nil {
		t.Fatalf("migrate category count tables: %v", err)
	}
	const base uint64 = 981000
	rows := []categoryCountTestTopic{
		{Id: base + 1},
		{Id: base + 2},
		{Id: base + 3},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}
	if err := conn.Delete(&rows[2]).Error; err != nil {
		t.Fatalf("soft delete topic: %v", err)
	}
	indexes := []Entity{
		{TopicId: base + 1, CategoryId: base + 11, Effective: 1},
		{TopicId: base + 1, CategoryId: base + 12, Effective: 1},
		{TopicId: base + 2, CategoryId: base + 11, Effective: 1},
		{TopicId: base + 3, CategoryId: base + 11, Effective: 1},
		{TopicId: base + 3, CategoryId: base + 13, Effective: 1},
	}
	if err := conn.Create(&indexes).Error; err != nil {
		t.Fatalf("create indexes: %v", err)
	}
	counts, err := MultiCategoryTopicCountsWithDB(conn, []uint64{base + 11, base + 12, base + 13})
	if err != nil {
		t.Fatalf("MultiCategoryTopicCountsWithDB: %v", err)
	}
	if counts[base+11] != 1 || counts[base+12] != 1 || counts[base+13] != 0 {
		t.Fatalf("multi-category counts = %v", counts)
	}
	ids, complete, err := ActiveTopicIDsByCategoryWithDB(conn, base+12, 2)
	if err != nil || !complete || len(ids) != 1 || ids[0] != base+1 {
		t.Fatalf("sparse category ids=%v complete=%v err=%v", ids, complete, err)
	}
	ids, complete, err = ActiveTopicIDsByCategoryWithDB(conn, base+11, 2)
	if err != nil || complete || len(ids) != 2 {
		t.Fatalf("truncated category ids=%v complete=%v err=%v", ids, complete, err)
	}
}
