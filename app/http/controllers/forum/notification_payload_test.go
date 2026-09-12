package forum

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
)

func TestBuildNotificationPayloadUsesPostNumberURL(t *testing.T) {
	notification := &eventNotification.Entity{
		Payload: eventNotification.NotificationPayload{TopicId: 42, PostId: 99, PostNo: 7},
	}

	item := BuildNotificationPayload(notification)
	if item.Topic == nil || item.Topic.URL != "/p/post/42/7" {
		t.Fatalf("notification topic = %#v, want post number URL", item.Topic)
	}
}

func TestBuildNotificationPayloadFallsBackToTopicURL(t *testing.T) {
	notification := &eventNotification.Entity{
		Payload: eventNotification.NotificationPayload{TopicId: 42, PostId: 99},
	}
	item := BuildNotificationPayload(notification)
	if item.Topic == nil || item.Topic.URL != "/p/post/42" {
		t.Fatalf("notification topic = %#v, want topic URL", item.Topic)
	}
}
