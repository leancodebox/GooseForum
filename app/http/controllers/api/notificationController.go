package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	forumcontroller "github.com/leancodebox/GooseForum/app/http/controllers/forum"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/service/notificationservice"
	"github.com/leancodebox/GooseForum/app/service/unreadservice"
)

// GetUnreadCountReq 获取未读数量请求
type GetUnreadCountReq struct{}

func GetUnreadStatus(req component.BetterRequest[GetUnreadCountReq]) component.Response {
	status := unreadservice.GetStatus(req.UserId)
	return component.SuccessResponse(component.DataMap{
		"notifications":          status.Notifications,
		"messages":               status.Messages,
		"latestNotificationType": status.LatestNotificationType,
	})
}

type NotificationListReq struct {
	Filter string `form:"filter"`
	Cursor uint64 `form:"cursor"`
	Limit  int    `form:"limit"`
}

type NotificationListResp struct {
	Items       []forumcontroller.NotificationPayload `json:"items"`
	NextCursor  uint64                                `json:"nextCursor"`
	HasNext     bool                                  `json:"hasNext"`
	UnreadCount int64                                 `json:"unreadCount"`
}

func NotificationList(req component.BetterRequest[NotificationListReq]) component.Response {
	unreadOnly := false
	switch req.Params.Filter {
	case "", "all":
	case "unread":
		unreadOnly = true
	default:
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}

	notifications, nextCursor, hasNext, err := notificationservice.GetNotificationCursorList(
		req.UserId,
		req.Params.Limit,
		req.Params.Cursor,
		unreadOnly,
	)
	if err != nil {
		return component.FailResponseCode(component.MessageRequestParseFailed, component.MessageParams{"error": err.Error()})
	}

	items := make([]forumcontroller.NotificationPayload, 0, len(notifications))
	for _, notification := range notifications {
		if notification == nil {
			continue
		}
		items = append(items, forumcontroller.BuildNotificationPayload(notification))
	}
	unreadCount, _ := eventNotification.GetUnreadCount(req.UserId)
	return component.SuccessResponse(NotificationListResp{
		Items:       items,
		NextCursor:  nextCursor,
		HasNext:     hasNext,
		UnreadCount: unreadCount,
	})
}

// MarkAsReadReq 标记通知已读请求
type MarkAsReadReq struct {
	NotificationId uint64 `json:"notificationId" validate:"required"`
}

// MarkAsRead 标记通知为已读
func MarkAsRead(req component.BetterRequest[MarkAsReadReq]) component.Response {
	err := eventNotification.MarkAsRead(req.Params.NotificationId, req.UserId)
	if err != nil {
		return component.FailResponseCode(component.MessageNotificationMarkReadFailed, nil)
	}
	unreadservice.Invalidate(req.UserId)

	return component.SuccessResponseCode("标记已读成功", component.MessageNotificationMarkReadSuccess, nil)
}

// MarkAllAsReadReq 标记所有通知已读请求
type MarkAllAsReadReq struct{}

// MarkAllAsRead 标记所有通知为已读
func MarkAllAsRead(req component.BetterRequest[MarkAllAsReadReq]) component.Response {
	err := eventNotification.MarkAllAsRead(req.UserId)
	if err != nil {
		return component.FailResponseCode(component.MessageNotificationMarkAllFailed, nil)
	}
	unreadservice.Invalidate(req.UserId)

	return component.SuccessResponseCode("标记全部已读成功", component.MessageNotificationMarkAllSuccess, nil)
}
