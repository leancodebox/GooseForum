package forum

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
)

func TestResolveUserProfileSection(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "default summary", raw: "", want: userProfileSectionSummary},
		{name: "activity", raw: "activity", want: userProfileSectionActivity},
		{name: "badges", raw: "badges", want: userProfileSectionBadges},
		{name: "unknown falls back", raw: "topics", want: userProfileSectionSummary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveUserProfileSection(tt.raw); got != tt.want {
				t.Fatalf("resolveUserProfileSection(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestResolveUserProfileActivitySection(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "default timeline", raw: "", want: userProfileActivityTimeline},
		{name: "topics", raw: "topics", want: userProfileActivityTopics},
		{name: "likes", raw: "likes", want: userProfileActivityLikes},
		{name: "following", raw: "following", want: userProfileActivityFollowing},
		{name: "followers", raw: "followers", want: userProfileActivityFollowers},
		{name: "unknown falls back", raw: "mentions", want: userProfileActivityTimeline},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveUserProfileActivitySection(tt.raw); got != tt.want {
				t.Fatalf("resolveUserProfileActivitySection(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestUserActivityURLForCommentUsesReplyArticle(t *testing.T) {
	activity := &userActivities.Entity{
		Action:      int(userActivities.ActionComment),
		SubjectType: userActivities.SubjectPost,
		SubjectId:   456,
	}
	replyByID := map[uint64]*reply.Entity{
		456: {Id: 456, ArticleId: 123},
	}

	if got := userActivityURL(activity, replyByID); got != "/p/post/123#reply-456" {
		t.Fatalf("userActivityURL() = %q, want %q", got, "/p/post/123#reply-456")
	}
}

func TestUserActivityURLForTopicUsesSubjectID(t *testing.T) {
	activity := &userActivities.Entity{
		Action:      int(userActivities.ActionPost),
		SubjectType: userActivities.SubjectTopic,
		SubjectId:   123,
	}

	if got := userActivityURL(activity, nil); got != "/p/post/123" {
		t.Fatalf("userActivityURL() = %q, want %q", got, "/p/post/123")
	}
}

func TestBuildUserActivityTopicCursorURL(t *testing.T) {
	if got := buildUserActivityTopicCursorURL(123, 456); got != "/u/123/activity/topics?cursor=456" {
		t.Fatalf("buildUserActivityTopicCursorURL() = %q", got)
	}
}

func TestBuildUserActivityTimelineCursorURL(t *testing.T) {
	if got := buildUserActivityTimelineCursorURL(123, 456); got != "/u/123/activity?cursor=456" {
		t.Fatalf("buildUserActivityTimelineCursorURL() = %q", got)
	}
}

func TestBuildUserActivityLikeCursorURL(t *testing.T) {
	if got := buildUserActivityLikeCursorURL(123, "456"); got != "/u/123/activity/likes?cursor=456" {
		t.Fatalf("buildUserActivityLikeCursorURL() = %q", got)
	}
}
