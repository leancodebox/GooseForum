package datamigration

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"gorm.io/gorm"
)

type UserActivityTopicMigrationResult struct {
	Topics     int64
	Posts      int64
	Missing    int64
	Failed     int
	LastFailed string
}

func BackfillUserActivityTopicIDs() UserActivityTopicMigrationResult {
	return BackfillUserActivityTopicIDsWithDB(dbconnect.Connect())
}

func BackfillUserActivityTopicIDsWithDB(conn *gorm.DB) UserActivityTopicMigrationResult {
	result := UserActivityTopicMigrationResult{}
	err := conn.Transaction(func(tx *gorm.DB) error {
		topicResult := tx.Model(&userActivities.Entity{}).
			Where("topic_id = 0 AND subject_type = ?", userActivities.SubjectTopic).
			Update("topic_id", gorm.Expr("subject_id"))
		if topicResult.Error != nil {
			return fmt.Errorf("backfill topic activities: %w", topicResult.Error)
		}
		result.Topics = topicResult.RowsAffected

		postResult := tx.Model(&userActivities.Entity{}).
			Where("topic_id = 0 AND subject_type = ?", userActivities.SubjectPost).
			Update("topic_id", gorm.Expr("COALESCE((SELECT topic_id FROM posts WHERE posts.id = user_activities.subject_id), 0)"))
		if postResult.Error != nil {
			return fmt.Errorf("backfill post activities: %w", postResult.Error)
		}
		result.Posts = postResult.RowsAffected
		if err := tx.Model(&userActivities.Entity{}).
			Where("topic_id = 0 AND subject_type IN ?", []string{userActivities.SubjectTopic, userActivities.SubjectPost}).
			Count(&result.Missing).Error; err != nil {
			return fmt.Errorf("count missing activity topics: %w", err)
		}
		return nil
	})
	if err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
	}
	return result
}
