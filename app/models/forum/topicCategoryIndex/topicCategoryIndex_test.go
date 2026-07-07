package topicCategoryIndex

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestTopicCategoryIndexRepositoryParity(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
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
