package articleUserAction

import (
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm/clause"
)

func GetByArticleId(userId, articleId any) (entity Entity) {
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldArticleId, articleId)).First(&entity)
	return
}

func SetLiked(userId, articleId uint64, liked bool) bool {
	return setAt(userId, articleId, fieldLikedAt, timeForState(liked))
}

func SetBookmarked(userId, articleId uint64, bookmarked bool) bool {
	return setAt(userId, articleId, fieldBookmarkedAt, timeForState(bookmarked))
}

func SetWatched(userId, articleId uint64, watched bool) bool {
	return setAt(userId, articleId, fieldWatchedAt, timeForState(watched))
}

func SetLikedAt(userId, articleId uint64, likedAt *time.Time) bool {
	return ensureAt(userId, articleId, fieldLikedAt, likedAt)
}

func SetBookmarkedAt(userId, articleId uint64, bookmarkedAt *time.Time) bool {
	return ensureAt(userId, articleId, fieldBookmarkedAt, bookmarkedAt)
}

func SetWatchedAt(userId, articleId uint64, watchedAt *time.Time) bool {
	return ensureAt(userId, articleId, fieldWatchedAt, watchedAt)
}

func setAt(userId, articleId uint64, field string, value *time.Time) bool {
	if userId == 0 || articleId == 0 {
		return false
	}
	if value == nil {
		result := builder().
			Where(queryopt.Eq(fieldUserId, userId)).
			Where(queryopt.Eq(fieldArticleId, articleId)).
			Where(field + " IS NOT NULL").
			Updates(map[string]any{field: nil, "updated_at": time.Now()})
		return result.Error == nil && result.RowsAffected > 0
	}

	result := builder().
		Where(queryopt.Eq(fieldUserId, userId)).
		Where(queryopt.Eq(fieldArticleId, articleId)).
		Where(field + " IS NULL").
		Updates(map[string]any{field: value, "updated_at": time.Now()})
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error == nil && result.RowsAffected > 0
	}

	result = builder().Create(&Entity{
		UserId:       userId,
		ArticleId:    articleId,
		LikedAt:      valueForField(field, fieldLikedAt, value),
		BookmarkedAt: valueForField(field, fieldBookmarkedAt, value),
		WatchedAt:    valueForField(field, fieldWatchedAt, value),
	})
	return result.Error == nil
}

func ensureAt(userId, articleId uint64, field string, value *time.Time) bool {
	if userId == 0 || articleId == 0 || value == nil {
		return false
	}
	result := builder().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: fieldUserId}, {Name: fieldArticleId}},
		DoUpdates: clause.Assignments(map[string]any{field: value}),
	}).Create(&Entity{
		UserId:       userId,
		ArticleId:    articleId,
		LikedAt:      valueForField(field, fieldLikedAt, value),
		BookmarkedAt: valueForField(field, fieldBookmarkedAt, value),
		WatchedAt:    valueForField(field, fieldWatchedAt, value),
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

func ListActiveWatchUserIDsAfter(articleId, afterUserId uint64, excludeUserIds []uint64, limit int) []uint64 {
	if articleId == 0 || limit <= 0 {
		return nil
	}

	userIds := make([]uint64, 0, limit)
	query := builder().
		Select(fieldUserId).
		Where(queryopt.Eq(fieldArticleId, articleId)).
		Where(fieldWatchedAt+" IS NOT NULL").
		Where(fieldUserId+" > ?", afterUserId)
	if len(excludeUserIds) > 0 {
		query = query.Where(fieldUserId+" NOT IN ?", excludeUserIds)
	}
	query.Order(fieldUserId+" ASC").Limit(limit).Pluck(fieldUserId, &userIds)
	return userIds
}

type LikedArticleRef struct {
	ID        uint64    `gorm:"column:id"`
	ArticleID uint64    `gorm:"column:article_id"`
	LikedAt   time.Time `gorm:"column:liked_at"`
}

func ListLikedArticleRefsBefore(userId uint64, cursor string, limit int) ([]LikedArticleRef, string) {
	if userId == 0 || limit <= 0 {
		return nil, ""
	}

	rows := make([]LikedArticleRef, 0, limit+1)
	cursorID := parseLikedCursor(cursor)
	query := builder().
		Select("id", fieldArticleId, fieldLikedAt).
		Where(queryopt.Eq(fieldUserId, userId)).
		Where(fieldLikedAt + " IS NOT NULL")
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
