package eventNotification

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
	"time"
)

// Create 创建通知
func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

// GetByUserId 获取用户的通知列表
func GetByUserId(userId uint64, limit, offset int, unreadOnly bool) (notifications []*Entity, total int64, err error) {
	db := builder().Where(queryopt.Eq("user_id", userId))

	if unreadOnly {
		db = db.Where(queryopt.Eq("is_read", false))
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	err = db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
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
		Updates(map[string]interface{}{
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
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// DeleteNotification 删除通知
func DeleteNotification(notificationId uint64, userId uint64) error {
	return builder().
		Where(queryopt.Eq("id", notificationId)).
		Where(queryopt.Eq("user_id", userId)). // 确保只能删除自己的通知
		Delete(&Entity{}).Error
}
