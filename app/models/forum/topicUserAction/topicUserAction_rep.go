package topicUserAction

import (
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
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

func GetByTopicId(userId, topicId any) (entity Entity) {
	builder().Where(queryopt.Eq("user_id", userId)).Where(queryopt.Eq("topic_id", topicId)).First(&entity)
	return
}

func SetLiked(userId, topicId uint64, liked bool) bool {
	return setAt(userId, topicId, "liked_at", timeForState(liked))
}

func SetBookmarked(userId, topicId uint64, bookmarked bool) bool {
	return setAt(userId, topicId, "bookmarked_at", timeForState(bookmarked))
}

func SetWatched(userId, topicId uint64, watched bool) bool {
	return setAt(userId, topicId, "watched_at", timeForState(watched))
}

func SetLikedAt(userId, topicId uint64, likedAt *time.Time) bool {
	return ensureAt(userId, topicId, "liked_at", likedAt)
}

func SetBookmarkedAt(userId, topicId uint64, bookmarkedAt *time.Time) bool {
	return ensureAt(userId, topicId, "bookmarked_at", bookmarkedAt)
}

func SetWatchedAt(userId, topicId uint64, watchedAt *time.Time) bool {
	return ensureAt(userId, topicId, "watched_at", watchedAt)
}

func setAt(userId, topicId uint64, field string, value *time.Time) bool {
	if userId == 0 || topicId == 0 {
		return false
	}
	if value == nil {
		result := builder().
			Where(queryopt.Eq("user_id", userId)).
			Where(queryopt.Eq("topic_id", topicId)).
			Where(field + " IS NOT NULL").
			Updates(map[string]any{field: nil, "updated_at": time.Now()})
		return result.Error == nil && result.RowsAffected > 0
	}

	result := builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("topic_id", topicId)).
		Where(field + " IS NULL").
		Updates(map[string]any{field: value, "updated_at": time.Now()})
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error == nil && result.RowsAffected > 0
	}

	result = builder().Create(&Entity{
		UserId:       userId,
		TopicId:      topicId,
		LikedAt:      valueForField(field, "liked_at", value),
		BookmarkedAt: valueForField(field, "bookmarked_at", value),
		WatchedAt:    valueForField(field, "watched_at", value),
	})
	return result.Error == nil
}

func ensureAt(userId, topicId uint64, field string, value *time.Time) bool {
	if userId == 0 || topicId == 0 || value == nil {
		return false
	}
	result := builder().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "topic_id"}},
		DoUpdates: clause.Assignments(map[string]any{field: value}),
	}).Create(&Entity{
		UserId:       userId,
		TopicId:      topicId,
		LikedAt:      valueForField(field, "liked_at", value),
		BookmarkedAt: valueForField(field, "bookmarked_at", value),
		WatchedAt:    valueForField(field, "watched_at", value),
	})
	return result.Error == nil
}

func timeForState(active bool) *time.Time {
	if !active {
		return nil
	}
	now := time.Now()
	return &now
}

func valueForField(field, target string, value *time.Time) *time.Time {
	if field != target {
		return nil
	}
	return value
}

func ListActiveWatchUserIDsAfter(topicId, afterUserId uint64, excludeUserIds []uint64, limit int) []uint64 {
	if topicId == 0 || limit <= 0 {
		return nil
	}

	userIds := make([]uint64, 0, limit)
	query := builder().
		Select("user_id").
		Where(queryopt.Eq("topic_id", topicId)).
		Where("watched_at IS NOT NULL").
		Where("user_id > ?", afterUserId)
	if len(excludeUserIds) > 0 {
		query = query.Where("user_id NOT IN ?", excludeUserIds)
	}
	query.Order("user_id ASC").Limit(limit).Pluck("user_id", &userIds)
	return userIds
}

type LikedTopicRef struct {
	ID      uint64    `gorm:"column:id"`
	TopicID uint64    `gorm:"column:topic_id"`
	LikedAt time.Time `gorm:"column:liked_at"`
}

func ListLikedTopicRefsBefore(userId uint64, cursor string, limit int) ([]LikedTopicRef, string) {
	if userId == 0 || limit <= 0 {
		return nil, ""
	}

	rows := make([]LikedTopicRef, 0, limit+1)
	cursorID := parseLikedCursor(cursor)
	query := builder().
		Select("id", "topic_id", "liked_at").
		Where(queryopt.Eq("user_id", userId)).
		Where("liked_at IS NOT NULL")
	if cursorID > 0 {
		query = query.Where("id < ?", cursorID)
	}
	query.Order("id DESC").Limit(limit + 1).Find(&rows)

	hasNext := len(rows) > limit
	if hasNext {
		rows = rows[:limit]
	}
	if hasNext && len(rows) > 0 {
		return rows, formatLikedCursor(rows[len(rows)-1].ID)
	}
	return rows, ""
}

func parseLikedCursor(cursor string) uint64 {
	id, err := strconv.ParseUint(cursor, 10, 64)
	if err != nil || id == 0 {
		return 0
	}
	return id
}

func formatLikedCursor(id uint64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatUint(id, 10)
}
