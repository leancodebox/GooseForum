package searchservice

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
)

func TestBuildExcerptKeepsChineseMatchesWithinSourcePost(t *testing.T) {
	first := "First post must not appear in reply excerpt.\n\n"
	reply := strings.Repeat("前文", 40) + "联动变化" + strings.Repeat("后文", 80)
	content := first + reply + "\n\nAnother reply must not appear."
	start := strings.Index(content, "联动变化")
	parts := buildExcerpt(content, SearchPost{Start: len(first), Length: len(reply)}, []matchPosition{
		{Start: start, Length: len("联动")},
		{Start: start + len("联动"), Length: len("变化")},
	})
	var text, highlighted strings.Builder
	for _, part := range parts {
		if !utf8.ValidString(part.Text) {
			t.Fatalf("invalid UTF-8 excerpt: %#v", parts)
		}
		text.WriteString(part.Text)
		if part.Matched {
			highlighted.WriteString(part.Text)
		}
	}
	if highlighted.String() != "联动变化" || strings.Contains(text.String(), "reply") || !strings.HasPrefix(text.String(), "...") || !strings.HasSuffix(text.String(), "...") {
		t.Fatalf("unexpected excerpt: %#v", parts)
	}
}

func TestBuildSearchResultsRejectsStaleAndHiddenPosts(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&posts.Entity{}); err != nil {
		t.Fatal(err)
	}
	const topicID = 980002
	conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
	t.Cleanup(func() { conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{}) })
	post := posts.Entity{TopicId: topicID, PostNo: 2, Content: "unique reply", ModerationVersion: 1, UpdatedAt: time.Now()}
	if err := conn.Create(&post).Error; err != nil {
		t.Fatal(err)
	}
	hit := searchHit{
		ID: topicID, Title: "Title", SearchContent: post.Content,
		SearchPosts:     []SearchPost{{PostID: post.Id, PostNo: 2, Version: 1, UpdatedAt: post.UpdatedAt.UnixMilli(), Length: len(post.Content)}},
		MatchesPosition: map[string][]matchPosition{"searchContent": {{Start: 0, Length: len("unique")}}},
	}
	check := func(wantMatch bool) {
		t.Helper()
		results, err := buildSearchResults([]searchHit{hit}, true)
		if err != nil || len(results) != 1 || (results[0].Match != nil) != wantMatch {
			t.Fatalf("results = %#v, err = %v, want match %v", results, err, wantMatch)
		}
	}
	check(true)
	if err := conn.Model(&post).Update("moderation_version", 2).Error; err != nil {
		t.Fatal(err)
	}
	check(false)
	hit.SearchPosts[0].Version = 2
	post = posts.Get(post.Id)
	hit.SearchPosts[0].UpdatedAt = post.UpdatedAt.UnixMilli()
	check(true)
	if err := conn.Model(&post).Update("process_status", 1).Error; err != nil {
		t.Fatal(err)
	}
	check(false)
	if err := conn.Delete(&post).Error; err != nil {
		t.Fatal(err)
	}
	check(false)
}
