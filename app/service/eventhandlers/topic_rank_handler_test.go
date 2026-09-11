package eventhandlers

import (
	"context"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestTopicRankEventRegistrationAndHandling(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&topics.Entity{}, &topicrank.Entity{}); err != nil {
		t.Fatal(err)
	}
	topic := topics.Entity{Status: 1, Title: "ranking event test"}
	if err := db.Create(&topic).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&topic); db.Delete(&topicrank.Entity{}, "topic_id = ?", topic.Id) })
	occurred := time.Now().Add(-time.Minute).Truncate(time.Millisecond)
	event := &topicrank.TopicRankRequested{TopicID: topic.Id, OccurredAt: occurred}
	found := false
	for _, handler := range Handlers() {
		if _, ok := handler.NewEvent().(*topicrank.TopicRankRequested); !ok {
			continue
		}
		found = true
		if err := handler.Handle(context.Background(), event); err != nil {
			t.Fatal(err)
		}
	}
	if !found {
		t.Fatal("ranking event handler is not registered")
	}
	var got topics.Entity
	if err := db.First(&got, topic.Id).Error; err != nil {
		t.Fatal(err)
	}
	scheduled, err := topicrank.Get(context.Background(), topic.Id)
	if err != nil {
		t.Fatal(err)
	}
	if scheduled.NextRunAt == nil || !scheduled.NextRunAt.Equal(occurred) || scheduled.Version != 1 {
		t.Fatalf("event schedule not preserved: %+v", scheduled)
	}
	if got.PublishedAt != nil || !got.UpdatedAt.Equal(topic.UpdatedAt) {
		t.Fatal("scheduling changed topic content")
	}

}
