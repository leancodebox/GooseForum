package posts

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func TestGetSearchReplies(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	const topicID = 980001
	conn.Unscoped().Where("topic_id IN ?", []uint64{topicID, topicID + 1}).Delete(&Entity{})
	t.Cleanup(func() {
		conn.Unscoped().Where("topic_id IN ?", []uint64{topicID, topicID + 1}).Delete(&Entity{})
	})
	rows := []Entity{
		{TopicId: topicID, PostNo: 1, Content: "first"},
		{TopicId: topicID, PostNo: 2, Content: "visible"},
		{TopicId: topicID, PostNo: 3, Content: "pending", ProcessStatus: 1, ModerationStatus: "pending"},
		{TopicId: topicID, PostNo: 4, Content: "blocked", ProcessStatus: 1},
		{TopicId: topicID, PostNo: 5, Content: "deleted"},
		{TopicId: topicID + 1, PostNo: 2, Content: "other topic"},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Delete(&rows[4]).Error; err != nil {
		t.Fatal(err)
	}
	assertReplies := func(want string) {
		t.Helper()
		got, err := GetSearchReplies(topicID)
		if err != nil || len(got) != 1 || got[0].Content != want {
			t.Fatalf("search replies = %#v, err = %v, want %q", got, err, want)
		}
	}
	assertReplies("visible")
	if err := conn.Model(&rows[1]).Updates(map[string]any{"content": "edited", "updated_at": time.Unix(200, 0)}).Error; err != nil {
		t.Fatal(err)
	}
	assertReplies("edited")
	if err := conn.Model(&rows[1]).Update("process_status", 1).Error; err != nil {
		t.Fatal(err)
	}
	got, err := GetSearchReplies(topicID)
	if err != nil || len(got) != 0 {
		t.Fatalf("blocked search replies = %#v, err = %v", got, err)
	}
}
