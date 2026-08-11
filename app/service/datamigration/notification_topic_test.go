package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"gorm.io/gorm"
)

func TestBackfillNotificationTopicIDsUsesPayloadAndIsIdempotent(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&eventNotification.Entity{}); err != nil {
		t.Fatalf("migrate notifications: %v", err)
	}
	rows := []eventNotification.Entity{
		{Id: 1, UserId: 1, EventType: eventNotification.EventTypeComment, Payload: eventNotification.NotificationPayload{TopicId: 42}},
		{Id: 2, UserId: 1, TopicId: 99, EventType: eventNotification.EventTypeComment, Payload: eventNotification.NotificationPayload{TopicId: 42}},
		{Id: 3, UserId: 1, EventType: eventNotification.EventTypeSystem},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatalf("create notifications: %v", err)
	}
	result := BackfillNotificationTopicIDsWithDB(conn)
	if result.Failed != 0 || result.Updated != 1 {
		t.Fatalf("first backfill = %+v", result)
	}
	result = BackfillNotificationTopicIDsWithDB(conn)
	if result.Failed != 0 || result.Updated != 0 {
		t.Fatalf("second backfill = %+v", result)
	}
	var stored []eventNotification.Entity
	if err := conn.Order("id ASC").Find(&stored).Error; err != nil {
		t.Fatalf("load notifications: %v", err)
	}
	if stored[0].TopicId != 42 || stored[1].TopicId != 99 || stored[2].TopicId != 0 {
		t.Fatalf("backfilled topic ids = [%d %d %d]", stored[0].TopicId, stored[1].TopicId, stored[2].TopicId)
	}
}
