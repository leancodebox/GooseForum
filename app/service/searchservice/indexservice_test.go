package searchservice

import (
	"strings"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestConvertTopicToSearchDocument(t *testing.T) {
	createdAt := time.Unix(1700000000, 0)
	updatedAt := time.Unix(1700000300, 0)
	topic := &topics.Entity{
		Id:            42,
		Title:         "Searchable title",
		CategoryIds:   []uint64{3, 5},
		Status:        1,
		ProcessStatus: 0,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
	firstPost := &posts.Entity{Content: "# Heading\n\nVisible text with [link](https://example.com).\n\n```go\nhidden()\n```"}

	got := convertTopicToSearchDocument(topic, firstPost)

	if got.ID != topic.Id || got.Title != topic.Title {
		t.Fatalf("unexpected identity fields: %#v", got)
	}
	if got.Type != 0 || got.TopicStatus != topic.Status || got.ProcessStatus != topic.ProcessStatus {
		t.Fatalf("unexpected status fields: %#v", got)
	}
	if got.CreatedAt != createdAt.Unix() || got.UpdatedAt != updatedAt.Unix() {
		t.Fatalf("unexpected timestamps: %#v", got)
	}
	if len(got.Category) != 2 || got.Category[0] != 3 || got.Category[1] != 5 {
		t.Fatalf("Category = %#v, want [3 5]", got.Category)
	}
	if !strings.Contains(got.SearchContent, "Visible text") {
		t.Fatalf("SearchContent should include readable text, got %q", got.SearchContent)
	}
	if strings.Contains(got.SearchContent, "hidden") {
		t.Fatalf("SearchContent should skip fenced code, got %q", got.SearchContent)
	}
}

func TestGetTaskUIDNil(t *testing.T) {
	if got := getTaskUID(nil); got != nil {
		t.Fatalf("getTaskUID(nil) = %v, want nil", got)
	}
}
