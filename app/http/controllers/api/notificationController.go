package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
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

// MarkAsReadReq 标记通知已读请求
type MarkAsReadReq struct {
	NotificationId uint64 `json:"notificationId" validate:"required"`
}

// MarkAsRead 标记通知为已读
func MarkAsRead(req component.BetterRequest[MarkAsReadReq]) component.Response {
	err := eventNotification.MarkAsRead(req.Params.NotificationId, req.UserId)
	if err != nil {
		return component.FailResponse("标记已读失败")
	}
	unreadservice.Invalidate(req.UserId)

	return component.SuccessResponse("标记已读成功")
}

// MarkAllAsReadReq 标记所有通知已读请求
type MarkAllAsReadReq struct{}

// MarkAllAsRead 标记所有通知为已读
func MarkAllAsRead(req component.BetterRequest[MarkAllAsReadReq]) component.Response {
	err := eventNotification.MarkAllAsRead(req.UserId)
	if err != nil {
		return component.FailResponse("标记全部已读失败")
	}
	unreadservice.Invalidate(req.UserId)

	return component.SuccessResponse("标记全部已读成功")
}
