package datamigration

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/meilisearch/meilisearch-go"
)

type TopicSearchIndexMigrationResult struct {
	Skipped                bool   `json:"skipped"`
	Rebuilt                bool   `json:"rebuilt"`
	ProcessedCount         int    `json:"processedCount"`
	FailedCount            int    `json:"failedCount"`
	LegacyIndexDeleteTried bool   `json:"legacyIndexDeleteTried"`
	LegacyIndexDeleted     bool   `json:"legacyIndexDeleted"`
	Failed                 int    `json:"failed"`
	LastFailed             string `json:"lastFailed"`
}

func MigrateTopicSearchIndex() TopicSearchIndexMigrationResult {
	return migrateTopicSearchIndex(
		meiliconnect.IsAvailable,
		searchservice.BuildMeilisearchIndex,
		func(index string) error {
			_, err := meiliconnect.GetClient().DeleteIndex(index)
			return err
		},
	)
}

func migrateTopicSearchIndex(
	isAvailable func() bool,
	buildIndex func() (*searchservice.IndexBuildResult, error),
	deleteIndex func(string) error,
) TopicSearchIndexMigrationResult {
	const legacyTopicIndex = "articles"

	result := TopicSearchIndexMigrationResult{}
	if !isAvailable() {
		result.Skipped = true
		return result
	}

	buildResult, err := buildIndex()
	if err != nil {
		result.Failed++
		result.LastFailed = err.Error()
		return result
	}
	result.Rebuilt = true
	result.ProcessedCount = buildResult.ProcessedCount
	result.FailedCount = buildResult.FailedCount

	if buildResult.FailedCount > 0 || legacyTopicIndex == "" || legacyTopicIndex == searchservice.TopicIndex {
		return result
	}
	result.LegacyIndexDeleteTried = true
	if err := deleteIndex(legacyTopicIndex); err != nil {
		if isMeiliNotFound(err) {
			result.LegacyIndexDeleted = true
			return result
		}
		slog.Warn("failed to delete legacy topic search index", "index", legacyTopicIndex, "err", err)
		return result
	}
	result.LegacyIndexDeleted = true
	return result
}

func isMeiliNotFound(err error) bool {
	var meiliErr *meilisearch.Error
	return errors.As(err, &meiliErr) && meiliErr.StatusCode == http.StatusNotFound
}
