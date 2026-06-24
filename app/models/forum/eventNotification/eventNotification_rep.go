package eventNotification

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
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
	db := builder().Where(queryopt.Eq("user_id", userId))
	if startId != 0 {
		db = db.Where(queryopt.Lt("id", startId))
	}
	if unreadOnly {
		db = db.Where(queryopt.Eq("is_read", false))
	}
	err = db.Order(queryopt.Desc(`id`)).
		Limit(limit).
		Find(&notifications).Error
	return
}

// GetLastUnread 获取用户未读通知数量
func GetLastUnread(userId uint64) (entity Entity) {
	builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("is_read", false)).
		Order("id DESC").
		First(&entity)
	return
}

// GetUnreadCount 获取用户未读通知数量
func GetUnreadCount(userId uint64) (count int64, err error) {
	err = builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("is_read", false)).
		Count(&count).Error
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
