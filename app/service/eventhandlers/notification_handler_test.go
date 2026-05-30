package eventhandlers

import "testing"

func TestCommentNotificationExcludeUserIds(t *testing.T) {
	event := &CommentCreatedEvent{
		UserId:              1,
		ArticleAuthorId:     2,
		ParentReplyAuthorId: 2,
	}

	userIds := commentNotificationExcludeUserIds(event)
	got := make(map[uint64]bool, len(userIds))
	for _, userId := range userIds {
		got[userId] = true
	}

	if len(got) != 2 || !got[1] || !got[2] {
		t.Fatalf("unexpected exclude user ids: %#v", userIds)
	}
}
