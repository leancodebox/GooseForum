package eventnotice

import (
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
)

func TestCommentNotificationsUseTopicPostPayload(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := db.Connect()
	if err := conn.AutoMigrate(&eventNotification.Entity{}); err != nil {
		t.Fatalf("migrate notifications: %v", err)
	}

	if err := SendCommentNotification(1, 10, "hello", 2, 99); err != nil {
		t.Fatalf("SendCommentNotification() err=%v", err)
	}

	var notification eventNotification.Entity
	if err := conn.First(&notification).Error; err != nil {
		t.Fatalf("load notification: %v", err)
	}
	if notification.Payload.TopicId != 10 || notification.Payload.PostId != 99 {
		t.Fatalf("payload topic/post = %d/%d, want 10/99", notification.Payload.TopicId, notification.Payload.PostId)
	}
	if notification.Payload.ArticleId != 0 || notification.Payload.CommentId != 0 {
		t.Fatalf("legacy payload article/comment = %d/%d, want zero", notification.Payload.ArticleId, notification.Payload.CommentId)
	}
}
