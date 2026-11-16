package searchservice

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/meilisearch/meilisearch-go"
	"gorm.io/gorm"
)

func TestUnpublishedTopicDeletesOnlyItsSearchDocument(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"taskUid":42,"indexUid":"topics","status":"enqueued","type":"documentDeletion"}`)
	}))
	defer server.Close()
	index := meilisearch.New(server.URL).Index(TopicIndex)
	task, err := syncTopicSearchDocument(index, &topics.Entity{Id: 123, Status: 0}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if task == nil || task.TaskUID != 42 {
		t.Fatalf("unexpected deletion task: %#v", task)
	}
	want := []string{"DELETE /indexes/" + TopicIndex + "/documents/123"}
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
}

func TestConfigureIndexUsesSettingsEndpoints(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"taskUid":1,"status":"enqueued","type":"settingsUpdate"}`)
	}))
	defer server.Close()
	if err := configureIndex(meilisearch.New(server.URL).Index(TopicIndex)); err != nil {
		t.Fatal(err)
	}
	prefix := "PUT /indexes/" + TopicIndex + "/settings/"
	want := []string{prefix + "searchable-attributes", prefix + "filterable-attributes", prefix + "sortable-attributes", prefix + "displayed-attributes"}
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
}

func TestConvertTopicToSearchDocument(t *testing.T) {
	createdAt := time.Unix(1700000000, 0)
	updatedAt := time.Unix(1700000300, 0)
	topic := &topics.Entity{
		Id:             42,
		Title:          "Searchable title",
		CategoryIds:    []uint64{3, 5},
		MainCategoryId: 3,
		Status:         1,
		ProcessStatus:  0,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	firstPost := &posts.Entity{Content: "# Heading\n\nVisible text with [link](https://example.com).\n\n```go\nhidden()\n```"}

	got := convertTopicToSearchDocument(topic, firstPost, nil)

	if got.ID != topic.Id || got.Title != topic.Title {
		t.Fatalf("unexpected identity fields: %#v", got)
	}
	if got.TopicStatus != topic.Status || got.ProcessStatus != topic.ProcessStatus {
		t.Fatalf("unexpected status fields: %#v", got)
	}
	if got.CreatedAt != createdAt.Unix() || got.UpdatedAt != updatedAt.Unix() {
		t.Fatalf("unexpected timestamps: %#v", got)
	}
	if len(got.Category) != 2 || got.Category[0] != 3 || got.Category[1] != 5 {
		t.Fatalf("Category = %#v, want [3 5]", got.Category)
	}
	if got.MainCategory != 3 {
		t.Fatalf("MainCategory = %d, want 3", got.MainCategory)
	}
	if !strings.Contains(got.SearchContent, "Visible text") {
		t.Fatalf("SearchContent should include readable text, got %q", got.SearchContent)
	}
	if strings.Contains(got.SearchContent, "hidden") {
		t.Fatalf("SearchContent should skip fenced code, got %q", got.SearchContent)
	}
}

func TestTopicIndexUsesTopicName(t *testing.T) {
	if TopicIndex != "topics" {
		t.Fatalf("TopicIndex = %q, want topics", TopicIndex)
	}
}

func TestConvertTopicToSearchDocumentExcludesHiddenFirstPost(t *testing.T) {
	for _, firstPost := range []*posts.Entity{
		nil,
		{Id: 1, PostNo: 1, Content: "hidden first post", ProcessStatus: 1},
		{Id: 1, PostNo: 1, Content: "deleted first post", DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true}},
	} {
		topic := &topics.Entity{Id: 42, Title: "Visible topic"}
		replies := []posts.SearchSource{{Id: 2, PostNo: 2, Content: "Visible reply"}}
		got := convertTopicToSearchDocument(topic, firstPost, replies)
		if strings.Contains(got.SearchContent, "first post") || !strings.Contains(got.SearchContent, "Visible") {
			t.Fatalf("unexpected search content: %q", got.SearchContent)
		}
		if len(got.SearchPosts) != 1 || got.SearchPosts[0].PostID != 2 {
			t.Fatalf("unexpected search sources: %#v", got.SearchPosts)
		}
		if got.Title != topic.Title {
			t.Fatalf("title = %q, want %q", got.Title, topic.Title)
		}
	}
}

func TestConvertTopicToSearchDocumentIncludesReplies(t *testing.T) {
	topic := &topics.Entity{Id: 42, UpdatedAt: time.Unix(100, 0)}
	first := &posts.Entity{Content: "Topic body"}
	replies := []posts.SearchSource{
		{Content: "Reply with [unique keyword](https://example.com)", UpdatedAt: time.Unix(200, 0)},
		{Content: "Another reply\n\n```go\nhidden()\n```", UpdatedAt: time.Unix(150, 0)},
	}
	got := convertTopicToSearchDocument(topic, first, replies)
	searchContent := strings.Join(strings.Fields(got.SearchContent), " ")
	for _, text := range []string{"Topic body", "unique keyword", "Another reply"} {
		if !strings.Contains(searchContent, text) {
			t.Fatalf("search content missing %q: %q", text, got.SearchContent)
		}
	}
	if strings.Contains(got.SearchContent, "hidden") || got.UpdatedAt != 200 {
		t.Fatalf("unexpected reply document: %#v", got)
	}
}

func TestTopicSearchDocumentDoesNotExposeLegacyType(t *testing.T) {
	if _, ok := reflect.TypeFor[TopicSearchDocument]().FieldByName("Type"); ok {
		t.Fatalf("TopicSearchDocument should not expose legacy Type field")
	}
}

func TestGetTaskUIDNil(t *testing.T) {
	if got := getTaskUID(nil); got != nil {
		t.Fatalf("getTaskUID(nil) = %v, want nil", got)
	}
}
