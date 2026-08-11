package eventNotification

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestAudienceQueryFiltersBeforeNotificationLimitAndUnreadCount(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&Entity{}, &topics.Entity{}, &topicCategoryIndex.Entity{}); err != nil {
		t.Fatalf("migrate notification audience tables: %v", err)
	}
	const userID uint64 = 950001
	publicTopic := topics.Entity{Id: 950101, Title: "public", CategoryIds: []uint64{1}, Status: 1}
	privateTopic := topics.Entity{Id: 950102, Title: "private", CategoryIds: []uint64{2}, Status: 1}
	if err := conn.Create(&[]topics.Entity{publicTopic, privateTopic}).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}
	if err := conn.Create(&[]topicCategoryIndex.Entity{
		{TopicId: publicTopic.Id, CategoryId: 1, Effective: 1},
		{TopicId: privateTopic.Id, CategoryId: 2, Effective: 1},
	}).Error; err != nil {
		t.Fatalf("create topic indexes: %v", err)
	}
	if err := conn.Create(&[]Entity{
		{Id: 950201, UserId: userID, TopicId: 0, EventType: EventTypeSystem},
		{Id: 950202, UserId: userID, TopicId: publicTopic.Id, EventType: EventTypeComment},
		{Id: 950203, UserId: userID, TopicId: privateTopic.Id, EventType: EventTypeComment},
	}).Error; err != nil {
		t.Fatalf("create notifications: %v", err)
	}
	list, err := QueryByUserIdForAudience(userID, 2, 0, false, []uint64{1}, true)
	if err != nil {
		t.Fatalf("query notifications: %v", err)
	}
	if len(list) != 2 || list[0].Id != 950202 || list[1].Id != 950201 {
		t.Fatalf("filtered notifications = %#v", list)
	}
	count, err := GetUnreadCountForAudience(userID, []uint64{1}, true)
	if err != nil || count != 2 {
		t.Fatalf("filtered unread count = %d, %v", count, err)
	}
}
