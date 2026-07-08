package datamigration

import (
	"errors"
	"net/http"
	"testing"

	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/meilisearch/meilisearch-go"
)

func TestMigrateTopicSearchIndexSkipsWhenUnavailable(t *testing.T) {
	result := migrateTopicSearchIndex(
		func() bool { return false },
		func() (*searchservice.IndexBuildResult, error) {
			t.Fatal("buildIndex should not be called when unavailable")
			return nil, nil
		},
		func(string) error {
			t.Fatal("deleteIndex should not be called when unavailable")
			return nil
		},
	)
	if !result.Skipped {
		t.Fatalf("MigrateTopicSearchIndex().Skipped = false, want true when Meilisearch is unavailable")
	}
	if result.Failed != 0 {
		t.Fatalf("MigrateTopicSearchIndex().Failed = %d, want 0", result.Failed)
	}
}

func TestMigrateTopicSearchIndexRebuildsThenDeletesLegacyIndex(t *testing.T) {
	deletedIndex := ""
	result := migrateTopicSearchIndex(
		func() bool { return true },
		func() (*searchservice.IndexBuildResult, error) {
			return &searchservice.IndexBuildResult{ProcessedCount: 3, FailedCount: 0, IndexName: searchservice.TopicIndex}, nil
		},
		func(index string) error {
			deletedIndex = index
			return nil
		},
	)

	if !result.Rebuilt || result.ProcessedCount != 3 || result.Failed != 0 {
		t.Fatalf("unexpected migration result: %#v", result)
	}
	if !result.LegacyIndexDeleteTried || !result.LegacyIndexDeleted || deletedIndex != "articles" {
		t.Fatalf("legacy delete state = tried:%v deleted:%v index:%q", result.LegacyIndexDeleteTried, result.LegacyIndexDeleted, deletedIndex)
	}
}

func TestMigrateTopicSearchIndexKeepsLegacyIndexWhenRebuildHasFailures(t *testing.T) {
	result := migrateTopicSearchIndex(
		func() bool { return true },
		func() (*searchservice.IndexBuildResult, error) {
			return &searchservice.IndexBuildResult{ProcessedCount: 3, FailedCount: 1, IndexName: searchservice.TopicIndex}, nil
		},
		func(index string) error {
			t.Fatalf("deleteIndex(%q) should not be called when rebuild has failures", index)
			return nil
		},
	)

	if !result.Rebuilt || result.FailedCount != 1 {
		t.Fatalf("unexpected migration result: %#v", result)
	}
	if result.LegacyIndexDeleteTried || result.LegacyIndexDeleted {
		t.Fatalf("legacy index should be kept when rebuild has failures: %#v", result)
	}
}

func TestMigrateTopicSearchIndexTreatsMissingLegacyIndexAsDeleted(t *testing.T) {
	result := migrateTopicSearchIndex(
		func() bool { return true },
		func() (*searchservice.IndexBuildResult, error) {
			return &searchservice.IndexBuildResult{ProcessedCount: 3, FailedCount: 0, IndexName: searchservice.TopicIndex}, nil
		},
		func(index string) error {
			if index != "articles" {
				t.Fatalf("deleteIndex(%q), want articles", index)
			}
			return &meilisearch.Error{StatusCode: http.StatusNotFound}
		},
	)

	if !result.LegacyIndexDeleteTried || !result.LegacyIndexDeleted {
		t.Fatalf("missing legacy index should be treated as already deleted: %#v", result)
	}
	if result.Failed != 0 {
		t.Fatalf("missing legacy index should not fail migration: %#v", result)
	}
}

func TestMigrateTopicSearchIndexReportsRebuildError(t *testing.T) {
	result := migrateTopicSearchIndex(
		func() bool { return true },
		func() (*searchservice.IndexBuildResult, error) {
			return nil, errors.New("boom")
		},
		func(index string) error {
			t.Fatalf("deleteIndex(%q) should not be called when rebuild fails", index)
			return nil
		},
	)

	if result.Failed != 1 || result.LastFailed != "boom" {
		t.Fatalf("unexpected failure result: %#v", result)
	}
}
