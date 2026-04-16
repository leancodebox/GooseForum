package notificationservice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/samber/lo"
)

func GetNotificationItemList(userId uint64, pageSize, offset int, unreadOnly bool) (int64, []*eventNotification.Entity) {
	notifications, total, err := eventNotification.GetByUserId(userId, pageSize, offset, unreadOnly)
	if err != nil {
		return 0, nil
	}
	userIds := lo.FilterMap(notifications, func(n *eventNotification.Entity, _ int) (uint64, bool) {
		return n.Payload.ActorId, n.Payload.ActorId != 0
	})
	articleIds := lo.FilterMap(notifications, func(n *eventNotification.Entity, _ int) (uint64, bool) {
		return n.Payload.ArticleId, n.Payload.ArticleId != 0
	})
	userMap := users.GetMapByIds(userIds)
	articleMap := articles.GetMapByIds(articleIds)

	// 转换数据
	lo.ForEach(notifications, func(notification *eventNotification.Entity, _ int) {
		if userInfo, ok := userMap[notification.Payload.ActorId]; ok {
			notification.Payload.ActorName = userInfo.Username
		}
		if articleInfo, ok := articleMap[notification.Payload.ArticleId]; ok {
			notification.Payload.ArticleTitle = articleInfo.Title
		}
	})
	return total, notifications
}
