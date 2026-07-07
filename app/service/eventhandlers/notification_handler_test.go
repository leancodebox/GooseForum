package eventhandlers

import (
	"strings"
	"testing"
)

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

func TestCommentCreatedEventTopicPostIDs(t *testing.T) {
	event := &CommentCreatedEvent{
		ArticleId: 10,
		CommentId: 20,
		TopicId:   30,
		PostId:    40,
	}
	if got := event.topicID(); got != 30 {
		t.Fatalf("topicID() = %d, want 30", got)
	}
	if got := event.postID(); got != 40 {
		t.Fatalf("postID() = %d, want 40", got)
	}

	legacy := &CommentCreatedEvent{ArticleId: 10, CommentId: 20}
	if got := legacy.topicID(); got != 10 {
		t.Fatalf("legacy topicID() = %d, want 10", got)
	}
	if got := legacy.postID(); got != 20 {
		t.Fatalf("legacy postID() = %d, want 20", got)
	}
}

func TestShouldNotifyArticleAuthor(t *testing.T) {
	tests := []struct {
		name  string
		event *CommentCreatedEvent
		want  bool
	}{
		{
			name: "root comment notifies article author",
			event: &CommentCreatedEvent{
				UserId:          1,
				ArticleAuthorId: 2,
			},
			want: true,
		},
		{
			name: "article author own comment does not notify",
			event: &CommentCreatedEvent{
				UserId:          2,
				ArticleAuthorId: 2,
			},
			want: false,
		},
		{
			name: "reply to another user still notifies article author",
			event: &CommentCreatedEvent{
				UserId:              1,
				ArticleAuthorId:     2,
				ParentReplyId:       10,
				ParentReplyAuthorId: 3,
			},
			want: true,
		},
		{
			name: "reply to article author only sends reply notification",
			event: &CommentCreatedEvent{
				UserId:              1,
				ArticleAuthorId:     2,
				ParentReplyId:       10,
				ParentReplyAuthorId: 2,
			},
			want: false,
		},
		{
			name: "missing article author does not notify",
			event: &CommentCreatedEvent{
				UserId: 1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldNotifyArticleAuthor(tt.event); got != tt.want {
				t.Fatalf("shouldNotifyArticleAuthor() = %v, want %v", got, tt.want)
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
				ParentReplyAuthorId: 2,
			},
			want: false,
		},
		{
			name: "reply notifies parent reply author",
			event: &CommentCreatedEvent{
				UserId:              1,
				ParentReplyId:       10,
				ParentReplyAuthorId: 2,
			},
			want: true,
		},
		{
			name: "self reply does not notify",
			event: &CommentCreatedEvent{
				UserId:              1,
				ParentReplyId:       10,
				ParentReplyAuthorId: 1,
			},
			want: false,
		},
		{
			name: "missing parent reply author does not notify",
			event: &CommentCreatedEvent{
				UserId:        1,
				ParentReplyId: 10,
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
