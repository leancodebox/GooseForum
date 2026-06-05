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

func TestTakeUpTo64Chars(t *testing.T) {
	short := "短内容"
	if got := TakeUpTo64Chars(short); got != short {
		t.Fatalf("short content changed: %q", got)
	}

	long := ""
	for i := 0; i < 70; i++ {
		long += "鹅"
	}
	got := TakeUpTo64Chars(long)
	if len([]rune(got)) != 64 {
		t.Fatalf("TakeUpTo64Chars() length = %d, want 64", len([]rune(got)))
	}
}
