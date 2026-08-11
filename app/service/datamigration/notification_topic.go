package datamigration

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"gorm.io/gorm"
)

const notificationTopicBackfillBatchSize = 500

type NotificationTopicMigrationResult struct {
	Updated    int64
	Failed     int
	LastFailed string
}

func BackfillNotificationTopicIDs() NotificationTopicMigrationResult {
	return BackfillNotificationTopicIDsWithDB(dbconnect.Connect())
}

func BackfillNotificationTopicIDsWithDB(conn *gorm.DB) NotificationTopicMigrationResult {
	result := NotificationTopicMigrationResult{}
	lastID := uint64(0)
	for {
		var rows []eventNotification.Entity
		if err := conn.
			Where("id > ? AND topic_id = 0", lastID).
			Order("id ASC").
			Limit(notificationTopicBackfillBatchSize).
			Find(&rows).Error; err != nil {
			result.Failed = 1
			result.LastFailed = fmt.Sprintf("load notification batch: %v", err)
			return result
		}
		if len(rows) == 0 {
			return result
		}
		err := conn.Transaction(func(tx *gorm.DB) error {
			for _, row := range rows {
				if row.Payload.TopicId == 0 {
					continue
				}
				update := tx.Model(&eventNotification.Entity{}).
					Where("id = ? AND topic_id = 0", row.Id).
					Update("topic_id", row.Payload.TopicId)
				if update.Error != nil {
					return fmt.Errorf("update notification %d: %w", row.Id, update.Error)
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
		if len(rows) < notificationTopicBackfillBatchSize {
			return result
		}
	}
}
