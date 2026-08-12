package datamigration

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"gorm.io/gorm"
)

const topicMainCategoryBatchSize = 500

type TopicMainCategoryMigrationResult struct {
	Updated    int64
	Orphaned   int64
	Failed     int
	LastFailed string
}

type TopicMainCategorySearchMigrationResult struct {
	Skipped        bool
	Rebuilt        bool
	ProcessedCount int
	FailedCount    int
	Failed         int
	LastFailed     string
}

// BackfillTopicMainCategory copies the first non-zero entry of every topic's
// category list into the new main_category_id column. It is purely additive:
// no topic changes categories, no row is deleted, and re-running it is a no-op
// because only rows still at 0 are considered.
func BackfillTopicMainCategory() TopicMainCategoryMigrationResult {
	return BackfillTopicMainCategoryWithDB(dbconnect.Connect())
}

func BackfillTopicMainCategoryWithDB(conn *gorm.DB) TopicMainCategoryMigrationResult {
	result := TopicMainCategoryMigrationResult{}
	lastID := uint64(0)
	for {
		var rows []topics.Entity
		if err := conn.
			Where("id > ? AND (main_category_id = 0 OR main_category_id IS NULL)", lastID).
			Order("id ASC").
			Limit(topicMainCategoryBatchSize).
			Find(&rows).Error; err != nil {
			result.Failed = 1
			result.LastFailed = fmt.Sprintf("load topic batch: %v", err)
			return result
		}
		if len(rows) == 0 {
			return result
		}
		err := conn.Transaction(func(tx *gorm.DB) error {
			for _, row := range rows {
				mainCategoryID := firstCategoryID(row.CategoryIds)
				if mainCategoryID == 0 {
					// A topic with no categories at all predates the category
					// requirement. Leave it at 0 rather than inventing one: 0 is
					// readable by nobody, which is the fail-closed side.
					result.Orphaned++
					continue
				}
				update := tx.Model(&topics.Entity{}).
					Where("id = ?", row.Id).
					Update("main_category_id", mainCategoryID)
				if update.Error != nil {
					return fmt.Errorf("update topic %d: %w", row.Id, update.Error)
				}
				result.Updated += update.RowsAffected
			}
			return nil
		})
		if err != nil {
			result.Failed = 1
			result.LastFailed = err.Error()
			return result
		}
		lastID = rows[len(rows)-1].Id
		if len(rows) < topicMainCategoryBatchSize {
			return result
		}
	}
}

// RebuildTopicMainCategorySearchIndex upgrades existing search documents with
// the new mainCategory field. A configured but unavailable Meilisearch is a
// migration failure: advancing the version would otherwise leave search empty
// until an operator happens to rebuild the index manually.
func RebuildTopicMainCategorySearchIndex() TopicMainCategorySearchMigrationResult {
	return rebuildTopicMainCategorySearchIndex(
		meiliconnect.IsConfigured,
		meiliconnect.IsAvailable,
		searchservice.BuildMeilisearchIndex,
	)
}

func rebuildTopicMainCategorySearchIndex(
	isConfigured func() bool,
	isAvailable func() bool,
	buildIndex func() (*searchservice.IndexBuildResult, error),
) TopicMainCategorySearchMigrationResult {
	result := TopicMainCategorySearchMigrationResult{}
	if !isConfigured() {
		result.Skipped = true
		return result
	}
	if !isAvailable() {
		result.Failed = 1
		result.LastFailed = "Meilisearch is configured but unavailable"
		return result
	}
	built, err := buildIndex()
	if err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
		return result
	}
	if built == nil {
		result.Failed = 1
		result.LastFailed = "search index rebuild returned no result"
		return result
	}
	result.Rebuilt = true
	result.ProcessedCount = built.ProcessedCount
	result.FailedCount = built.FailedCount
	if built.FailedCount > 0 {
		result.Failed = 1
		result.LastFailed = fmt.Sprintf("%d topic search documents failed to rebuild", built.FailedCount)
	}
	return result
}

func firstCategoryID(categoryIDs []uint64) uint64 {
	for _, categoryID := range categoryIDs {
		if categoryID != 0 {
			return categoryID
		}
	}
	return 0
}
