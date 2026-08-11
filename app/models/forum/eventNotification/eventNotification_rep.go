package eventNotification

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm"
)

// Create 创建通知
func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func CreateBatch(entities []*Entity, batchSize int) error {
	if len(entities) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return builder().CreateInBatches(entities, batchSize).Error
}

// QueryByUserId 获取用户的通知列表
func QueryByUserId(userId uint64, limit int, startId uint64, unreadOnly bool) (notifications []*Entity, err error) {
	return QueryByUserIdForAudience(userId, limit, startId, unreadOnly, nil, false)
}

func QueryByUserIdForAudience(userID uint64, limit int, startID uint64, unreadOnly bool, readableCategoryIDs []uint64, filterAudience bool) (notifications []*Entity, err error) {
	db := builder().Where(queryopt.Eq("event_notification.user_id", userID))
	if startID != 0 {
		db = db.Where(queryopt.Lt("event_notification.id", startID))
	}
	if unreadOnly {
		db = db.Where(queryopt.Eq("event_notification.is_read", false))
	}
	db = applyAudienceFilter(db, readableCategoryIDs, filterAudience)
	err = db.Order(queryopt.Desc(`event_notification.id`)).
		Limit(limit).
		Find(&notifications).Error
	return
}

func applyAudienceFilter(db *gorm.DB, readableCategoryIDs []uint64, filterAudience bool) *gorm.DB {
	if !filterAudience {
		return db
	}
	if len(readableCategoryIDs) == 0 {
		return db.Where("event_notification.topic_id = 0")
	}
	return db.
		Joins("LEFT JOIN topics ON topics.id = event_notification.topic_id").
		Where(`event_notification.topic_id = 0 OR (
			topics.id IS NOT NULL
			AND topics.status = 1
			AND topics.process_status = 0
			AND EXISTS (
				SELECT 1 FROM topic_category_index audience_idx
				WHERE audience_idx.topic_id = topics.id
				AND audience_idx.effective = 1
				AND audience_idx.category_id IN ?
			)
		)`, readableCategoryIDs)
}

// GetLastUnread 获取用户未读通知数量
func GetLastUnread(userId uint64) (entity Entity) {
	return GetLastUnreadForAudience(userId, nil, false)
}

func GetLastUnreadForAudience(userID uint64, readableCategoryIDs []uint64, filterAudience bool) (entity Entity) {
	query := builder().
		Where(queryopt.Eq("event_notification.user_id", userID)).
		Where(queryopt.Eq("event_notification.is_read", false))
	query = applyAudienceFilter(query, readableCategoryIDs, filterAudience)
	query.Order("event_notification.id DESC").
		First(&entity)
	return
}

// GetUnreadCount 获取用户未读通知数量
func GetUnreadCount(userId uint64) (count int64, err error) {
	return GetUnreadCountForAudience(userId, nil, false)
}

func GetUnreadCountForAudience(userID uint64, readableCategoryIDs []uint64, filterAudience bool) (count int64, err error) {
	query := builder().
		Where(queryopt.Eq("event_notification.user_id", userID)).
		Where(queryopt.Eq("event_notification.is_read", false))
	query = applyAudienceFilter(query, readableCategoryIDs, filterAudience)
	err = query.Count(&count).Error
	return
}

// MarkAsRead 标记通知为已读
func MarkAsRead(notificationId uint64, userId uint64) error {
	now := time.Now()
	return builder().
		Where(queryopt.Eq("id", notificationId)).
		Where(queryopt.Eq("user_id", userId)). // 确保只能标记自己的通知
		Updates(map[string]any{
			"is_read": true,
			"read_at": now,
		}).Error
}

// MarkAllAsRead 标记用户所有通知为已读
func MarkAllAsRead(userId uint64) error {
	now := time.Now()
	return builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("is_read", false)).
		Updates(map[string]any{
			"is_read": true,
			"read_at": now,
		}).Error
}
