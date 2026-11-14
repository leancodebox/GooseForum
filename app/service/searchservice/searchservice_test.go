package searchservice

import "testing"

func TestBuildTopicFilterAlwaysExcludesHiddenTopics(t *testing.T) {
	if got := buildTopicFilter(SearchRequest{}); got != "topicStatus = 1 AND processStatus = 0" {
		t.Fatalf("filter = %q", got)
	}
}

func TestBuildTopicFilterAddsReadableCategories(t *testing.T) {
	got := buildTopicFilter(SearchRequest{
		Categories:         []uint64{2, 7},
		FilterByCategories: true,
	})
	want := "topicStatus = 1 AND processStatus = 0 AND (mainCategory = 2 OR mainCategory = 7)"
	if got != want {
		t.Fatalf("filter = %q, want %q", got, want)
	}
}
