package notificationservice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/samber/lo"
)

const (
	DefaultNotificationPageSize = 20
	MaxNotificationPageSize     = 50
)

func GetNotificationCursorList(userId uint64, pageSize int, cursor uint64, unreadOnly bool) ([]*eventNotification.Entity, uint64, bool, error) {
	pageSize = normalizePageSize(pageSize)
	notifications, err := eventNotification.QueryByUserId(userId, pageSize+1, cursor, unreadOnly)
	if err != nil {
		return nil, 0, false, err
	}

	hasNext := len(notifications) > pageSize
	if hasNext {
		notifications = notifications[:pageSize]
	}
	hydrateNotifications(notifications)

	nextCursor := uint64(0)
	if hasNext && len(notifications) > 0 {
		nextCursor = notifications[len(notifications)-1].Id
	}
	return notifications, nextCursor, hasNext, nil
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return DefaultNotificationPageSize
	}
	if pageSize > MaxNotificationPageSize {
		return MaxNotificationPageSize
	}
	return pageSize
}

func hydrateNotifications(notifications []*eventNotification.Entity) {
	userIds := lo.FilterMap(notifications, func(n *eventNotification.Entity, _ int) (uint64, bool) {
		return n.Payload.ActorId, n.Payload.ActorId != 0
	})
	topicIds := lo.FilterMap(notifications, func(n *eventNotification.Entity, _ int) (uint64, bool) {
		return n.Payload.TopicId, n.Payload.TopicId != 0
	})
	userMap := users.GetMapByIds(userIds)
	topicMap := topics.GetMapByIds(topicIds)

	// 转换数据
	lo.ForEach(notifications, func(notification *eventNotification.Entity, _ int) {
		if userInfo, ok := userMap[notification.Payload.ActorId]; ok {
			notification.Payload.ActorName = userInfo.Username
		}
		if topicInfo, ok := topicMap[notification.Payload.TopicId]; ok {
			notification.Payload.TopicTitle = topicInfo.Title
		}
	})
}
