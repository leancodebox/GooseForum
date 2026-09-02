package datamigration

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"gorm.io/gorm"
)

type CategoryTopicCountMigrationResult struct {
	Categories int64
	Failed     int
	LastFailed string
}

type categoryTopicCountRow struct {
	CategoryID uint64 `gorm:"column:category_id"`
	TopicCount uint64 `gorm:"column:topic_count"`
}

func BackfillCategoryTopicCounts() CategoryTopicCountMigrationResult {
	return BackfillCategoryTopicCountsWithDB(dbconnect.Connect())
}

func BackfillCategoryTopicCountsWithDB(conn *gorm.DB) CategoryTopicCountMigrationResult {
	result := CategoryTopicCountMigrationResult{}
	if conn == nil {
		result.Failed = 1
		result.LastFailed = "database is required"
		return result
	}
	if err := conn.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&category.Entity{}).UpdateColumn("topic_count", 0).Error; err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
		return result
	}
	var rows []categoryTopicCountRow
	err := conn.Table("topic_category_index AS category_idx").
		Select("category_idx.category_id AS category_id, COUNT(DISTINCT category_idx.topic_id) AS topic_count").
		Joins("JOIN topics ON topics.id = category_idx.topic_id AND topics.deleted_at IS NULL AND topics.status = ? AND topics.process_status = ?", 1, 0).
		Where("category_idx.effective = ?", 1).
		Group("category_idx.category_id").
		Scan(&rows).Error
	if err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
		return result
	}
	for _, row := range rows {
		update := conn.Model(&category.Entity{}).
			Where("id = ?", row.CategoryID).
			UpdateColumn("topic_count", row.TopicCount)
		if update.Error != nil {
			result.Failed++
			result.LastFailed = update.Error.Error()
			continue
		}
		result.Categories += update.RowsAffected
	}
	return result
}
