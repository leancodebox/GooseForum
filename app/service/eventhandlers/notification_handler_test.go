package eventhandlers

import (
	"strings"
	"testing"
)

func TestCommentNotificationExcludeUserIds(t *testing.T) {
	event := &CommentCreatedEvent{
		UserId:              1,
		TopicAuthorId:       2,
		ReplyToPostAuthorId: 2,
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

func TestCommentCreatedEventCarriesTopicPostIDs(t *testing.T) {
	event := &CommentCreatedEvent{TopicId: 30, PostId: 40}
	if event.TopicId != 30 || event.PostId != 40 {
		t.Fatalf("event ids = %#v", event)
	}
}

func TestShouldNotifyTopicAuthor(t *testing.T) {
	tests := []struct {
		name  string
		event *CommentCreatedEvent
		want  bool
	}{
		{
			name: "root comment notifies topic author",
			event: &CommentCreatedEvent{
				UserId:        1,
				TopicAuthorId: 2,
			},
			want: true,
		},
		{
			name: "topic author own comment does not notify",
			event: &CommentCreatedEvent{
				UserId:        2,
				TopicAuthorId: 2,
			},
			want: false,
		},
		{
			name: "reply to another user still notifies topic author",
			event: &CommentCreatedEvent{
				UserId:              1,
				TopicAuthorId:       2,
				ReplyToPostId:       10,
				ReplyToPostAuthorId: 3,
			},
			want: true,
		},
		{
			name: "reply to topic author only sends reply notification",
			event: &CommentCreatedEvent{
				UserId:              1,
				TopicAuthorId:       2,
				ReplyToPostId:       10,
				ReplyToPostAuthorId: 2,
			},
			want: false,
		},
		{
			name: "missing topic author does not notify",
			event: &CommentCreatedEvent{
				UserId: 1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldNotifyTopicAuthor(tt.event); got != tt.want {
				t.Fatalf("shouldNotifyTopicAuthor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldNotifyParentReplyAuthor(t *testing.T) {
	tests := []struct {
		name  string
		event *CommentCreatedEvent
		want  bool
	}{
		{
			name: "root comment does not notify parent reply author",
			event: &CommentCreatedEvent{
				UserId:              1,
				ReplyToPostAuthorId: 2,
			},
			want: false,
		},
		{
			name: "reply notifies parent reply author",
			event: &CommentCreatedEvent{
				UserId:              1,
				ReplyToPostId:       10,
				ReplyToPostAuthorId: 2,
			},
			want: true,
		},
		{
			name: "self reply does not notify",
			event: &CommentCreatedEvent{
				UserId:              1,
				ReplyToPostId:       10,
				ReplyToPostAuthorId: 1,
			},
			want: false,
		},
		{
			name: "missing parent reply author does not notify",
			event: &CommentCreatedEvent{
				UserId:        1,
				ReplyToPostId: 10,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldNotifyParentReplyAuthor(tt.event); got != tt.want {
				t.Fatalf("shouldNotifyParentReplyAuthor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTakeUpTo64Chars(t *testing.T) {
	short := "短内容"
	if got := TakeUpTo64Chars(short); got != short {
		t.Fatalf("short content changed: %q", got)
	}

	var long strings.Builder
	for range 70 {
		long.WriteString("鹅")
	}
	got := TakeUpTo64Chars(long.String())
	if len([]rune(got)) != 64 {
		t.Fatalf("TakeUpTo64Chars() length = %d, want 64", len([]rune(got)))
	}
}
