package forum

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestBuildCategoriesPagePropsPreservesVisibleOrderAndCounts(t *testing.T) {
	props := buildCategoriesPageProps([]*category.Entity{
		{Id: 8, Name: "General", Slug: "general", Color: "#123456"},
		nil,
		{Id: 3, Name: "Design", Slug: "design", Color: "#abcdef"},
	}, map[uint64]int64{8: 12, 3: 4})

	if props.Total != 2 || len(props.Categories) != 2 {
		t.Fatalf("categories total/list = %d/%d", props.Total, len(props.Categories))
	}
	if props.Categories[0].ID != 8 || props.Categories[0].TopicCount != 12 {
		t.Fatalf("first category = %#v", props.Categories[0])
	}
	if props.Categories[1].URL != "/c/design/3" {
		t.Fatalf("second category url = %q", props.Categories[1].URL)
	}
}

func TestBuildMembersPagePropsExposesOnlyDirectoryFields(t *testing.T) {
	joined := time.Date(2025, 7, 3, 0, 0, 0, 0, time.UTC)
	active := time.Date(2026, 9, 2, 12, 30, 0, 0, time.UTC)
	result := users.PublicPageResult{
		HasPrevious: true, HasNext: true,
		Data: []users.EntityComplete{{Id: 9, Username: "goose", Email: "private@example.com", CreatedAt: joined}},
	}
	props := buildMembersPageProps(result, map[uint64]*userStatistics.Entity{
		9: {UserId: 9, TopicCount: 3, ReplyCount: 11, LastActiveTime: active},
	})

	if props.PreviousURL != "/members?after=9" || !props.Pagination.HasNext || props.Pagination.NextURL != "/members?before=9" {
		t.Fatalf("members pagination = %#v", props)
	}
	member := props.Members[0]
	if member.Nickname != "goose" || member.TopicCount != 3 || member.ReplyCount != 11 {
		t.Fatalf("member = %#v", member)
	}
	if member.JoinedAt != "2025-07" || member.LastActiveAt != "2026-09-02 12:30:00" {
		t.Fatalf("member dates = %q/%q", member.JoinedAt, member.LastActiveAt)
	}
}

func TestMembersCursorURL(t *testing.T) {
	cases := map[string]struct {
		direction string
		id        uint64
		want      string
	}{
		"before":           {direction: "before", id: 42, want: "/members?before=42"},
		"after":            {direction: "after", id: 99, want: "/members?after=99"},
		"missingId":        {direction: "before", want: ""},
		"invalidDirection": {direction: "page", id: 2, want: ""},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			if got := membersCursorURL(test.direction, test.id); got != test.want {
				t.Fatalf("membersCursorURL(%q, %d) = %q, want %q", test.direction, test.id, got, test.want)
			}
		})
	}
}
