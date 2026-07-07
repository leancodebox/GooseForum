package topicUserStat

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return builder().Create(entity).RowsAffected
	}

	return builder().Save(entity).RowsAffected
}

func Get(id uint64) (entity Entity) {
	builder().First(&entity, id)
	return
}

func IncrementUserPost(topicId, userId uint64) error {
	now := time.Now()
	return builder().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "topic_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"reply_count":   gorm.Expr("reply_count + 1"),
			"last_reply_at": now,
		}),
	}).Create(map[string]any{
		"topic_id":      topicId,
		"user_id":       userId,
		"reply_count":   1,
		"last_reply_at": now,
	}).Error
}

func DecrementUserPost(topicId, userId uint64) error {
	return builder().
		Where("topic_id = ? AND user_id = ? AND reply_count > 0", topicId, userId).
		Update("reply_count", gorm.Expr("reply_count - 1")).
		Error
}

func SyncTopicPosters(topicId uint64) []uint64 {
	var activeUserIDs []uint64
	builder().
		Where("topic_id = ?", topicId).
		Order("reply_count DESC, last_reply_at DESC").
		Limit(3).
		Pluck("user_id", &activeUserIDs)
	return activeUserIDs
}
