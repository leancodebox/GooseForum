package topicUserStat

import (
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func TestTopicUserStatRepositoryParity(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate topic user stat: %v", err)
	}
	conn.Where("1 = 1").Delete(&Entity{})

	if err := IncrementUserPost(10, 1); err != nil {
		t.Fatalf("IncrementUserPost first err=%v", err)
	}
	if err := IncrementUserPost(10, 1); err != nil {
		t.Fatalf("IncrementUserPost second err=%v", err)
	}
	if err := IncrementUserPost(10, 2); err != nil {
		t.Fatalf("IncrementUserPost user 2 err=%v", err)
	}

	var row Entity
	conn.Where("topic_id = ? AND user_id = ?", 10, 1).First(&row)
	if row.ReplyCount != 2 {
		t.Fatalf("ReplyCount=%d, want 2", row.ReplyCount)
	}
	if got := SyncTopicPosters(10); !reflect.DeepEqual(got, []uint64{1, 2}) {
		t.Fatalf("SyncTopicPosters()=%#v, want [1 2]", got)
	}
	if err := DecrementUserPost(10, 1); err != nil {
		t.Fatalf("DecrementUserPost err=%v", err)
	}
	conn.Where("topic_id = ? AND user_id = ?", 10, 1).First(&row)
	if row.ReplyCount != 1 {
		t.Fatalf("ReplyCount after decrement=%d, want 1", row.ReplyCount)
	}
}
