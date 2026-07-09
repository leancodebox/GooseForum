package transform

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestTopicsWithUser2VoMapsListPayload(t *testing.T) {
	now := time.Date(2026, 7, 9, 12, 30, 0, 0, time.UTC)
	topic := &topics.Entity{
		Id:            10,
		Title:         "topic title",
		Excerpt:       "topic excerpt",
		FirstImageURL: "/cover.png",
		CategoryIds:   []uint64{3, 4},
		UserId:        1,
		ReplyCount:    7,
		ViewCount:     99,
		PinWeight:     5,
		ProcessStatus: 2,
		Posters:       []topics.Poster{{UserID: 1}, {UserID: 2}},
		CreatedAt:     now.Add(-time.Hour),
		UpdatedAt:     now,
	}
	categoryMap := map[uint64]*category.Entity{
		3: {Id: 3, Name: "General"},
	}
	userMap := map[uint64]*users.EntityComplete{
		1: {Id: 1, Username: "author", AvatarUrl: "/static/pic/author.webp"},
		2: {Id: 2, Username: "replyer", AvatarUrl: "/static/pic/replyer.webp"},
	}

	got := TopicsWithUser2Vo([]*topics.Entity{topic}, categoryMap, userMap)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	item := got[0]
	if item.Id != topic.Id || item.Title != topic.Title || item.Description != topic.Excerpt {
		t.Fatalf("basic fields not mapped: %#v", item)
	}
	if item.Username != "author" || item.AvatarUrl != "/static/pic/author.webp" {
		t.Fatalf("author fields not mapped: %#v", item)
	}
	if len(item.Categories) != 2 || item.Categories[0] != "General" || item.Categories[1] != "" {
		t.Fatalf("Categories = %#v, want General and fallback empty", item.Categories)
	}
	if len(item.Posters) != 2 || item.Posters[1].Username != "replyer" {
		t.Fatalf("Posters = %#v, want replyer poster", item.Posters)
	}
	if item.LastUpdateTime != now.Format(time.DateTime) {
		t.Fatalf("LastUpdateTime = %q, want %q", item.LastUpdateTime, now.Format(time.DateTime))
	}
}
