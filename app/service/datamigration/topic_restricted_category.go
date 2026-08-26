package datamigration

import (
	"fmt"
	"slices"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

type TopicRestrictedCategoryMigrationResult struct {
	Updated              int64
	RemovedCategoryLinks int64
	Failed               int
	LastFailed           string
}

// EnforceSingleRestrictedTopicCategory repairs selections created before the
// invariant existed without changing a topic's audience. If the main category
// is restricted, it is kept alone. If the main category is public, restricted
// auxiliary tags are removed while all public categories keep their order.
func EnforceSingleRestrictedTopicCategory() TopicRestrictedCategoryMigrationResult {
	return EnforceSingleRestrictedTopicCategoryWithDB(dbconnect.Connect())
}

func EnforceSingleRestrictedTopicCategoryWithDB(conn *gorm.DB) TopicRestrictedCategoryMigrationResult {
	result := TopicRestrictedCategoryMigrationResult{}
	if conn == nil {
		result.Failed = 1
		result.LastFailed = "database is unavailable"
		return result
	}
	public, err := loadPublicCategoryIDs(conn)
	if err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
		return result
	}
	lastID := uint64(0)
	for {
		var rows []topics.Entity
		if err := conn.Unscoped().
			Where("id > ?", lastID).
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
		var batchUpdated int64
		var batchRemovedCategoryLinks int64
		err := conn.Transaction(func(tx *gorm.DB) error {
			for _, row := range rows {
				current := canonicalTopicCategoryIDs(row.CategoryIds)
				next := singleRestrictedCategorySelection(current, public)
				if slices.Equal(current, next) {
					continue
				}
				if len(next) == 0 {
					continue
				}
				row.CategoryIds = next
				row.MainCategoryId = next[0]
				update := tx.Unscoped().Model(&row).
					Select("CategoryIds", "MainCategoryId").
					Omit("UpdatedAt").
					Updates(&row)
				if update.Error != nil {
					return fmt.Errorf("update topic %d: %w", row.Id, update.Error)
				}
				if err := topicCategoryIndex.ReplaceTopicCategoriesWithDB(tx, row.Id, next); err != nil {
					return fmt.Errorf("replace topic %d category indexes: %w", row.Id, err)
				}
				batchUpdated += update.RowsAffected
				batchRemovedCategoryLinks += int64(len(current) - len(next))
			}
			return nil
		})
		if err != nil {
			result.Failed = 1
			result.LastFailed = err.Error()
			return result
		}
		result.Updated += batchUpdated
		result.RemovedCategoryLinks += batchRemovedCategoryLinks
		lastID = rows[len(rows)-1].Id
		if len(rows) < topicMainCategoryBatchSize {
			return result
		}
	}
}

func loadPublicCategoryIDs(conn *gorm.DB) (map[uint64]struct{}, error) {
	var everyone accessGroups.Entity
	if err := conn.
		Where("system_key = ? AND status = ?", accessGroups.SystemKeyEveryone, accessGroups.StatusEnabled).
		First(&everyone).Error; err != nil {
		return nil, fmt.Errorf("load everyone access group: %w", err)
	}
	var grants []categoryGroupPermissions.Entity
	if err := conn.
		Where("access_group_id = ? AND status = ? AND permission_level >= ?", everyone.Id, categoryGroupPermissions.StatusEnabled, categoryGroupPermissions.PermissionRead).
		Find(&grants).Error; err != nil {
		return nil, fmt.Errorf("load public category grants: %w", err)
	}
	public := make(map[uint64]struct{}, len(grants))
	for _, grant := range grants {
		public[grant.CategoryId] = struct{}{}
	}
	return public, nil
}

func singleRestrictedCategorySelection(categoryIDs []uint64, public map[uint64]struct{}) []uint64 {
	if len(categoryIDs) <= 1 {
		return categoryIDs
	}
	if _, mainIsPublic := public[categoryIDs[0]]; !mainIsPublic {
		return []uint64{categoryIDs[0]}
	}
	next := make([]uint64, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		if _, isPublic := public[categoryID]; isPublic {
			next = append(next, categoryID)
		}
	}
	return next
}

func canonicalTopicCategoryIDs(categoryIDs []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(categoryIDs))
	result := make([]uint64, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		if categoryID == 0 {
			continue
		}
		if _, ok := seen[categoryID]; ok {
			continue
		}
		seen[categoryID] = struct{}{}
		result = append(result, categoryID)
	}
	return result
}
