package controllers

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

// GetNotificationListReq 获取通知列表请求
type GetNotificationListReq struct {
	Page     int `json:"page" validate:"required,min=1"`
	PageSize int `json:"pageSize" validate:"required,min=1,max=100"`
}

// GetNotificationList 获取通知列表
func GetNotificationList(req component.BetterRequest[GetNotificationListReq]) component.Response {
	offset := (req.Params.Page - 1) * req.Params.PageSize
	notifications, total, err := eventNotification.GetByUserId(req.UserId, req.Params.PageSize, offset)
	if err != nil {
		return component.FailResponse("获取通知列表失败")
	}

	// 转换数据
	for _, notification := range notifications {
		fmt.Println(notification.Payload.ActorId)
		if notification.Payload.ActorId != 0 {
			userInfo, err := users.Get(notification.Payload.ActorId)
			if err == nil {
				notification.Payload.ActorName = userInfo.Username
			}
		}
		fmt.Println(notification.Payload.ArticleId)
		if notification.Payload.ArticleId != 0 {
			articleInfo := articles.Get(notification.Payload.ArticleId)
			if articleInfo.Id != 0 {
				notification.Payload.ArticleTitle = articleInfo.Title
			}
		}
	}

	return component.SuccessResponse(component.DataMap{
		"list":  notifications,
		"total": total,
		"page":  req.Params.Page,
	})
}

// GetUnreadCountReq 获取未读数量请求
type GetUnreadCountReq struct{}

// GetUnreadCount 获取未读通知数量
func GetUnreadCount(req component.BetterRequest[GetUnreadCountReq]) component.Response {
	count, err := eventNotification.GetUnreadCount(req.UserId)
	if err != nil {
		return component.FailResponse("获取未读数量失败")
	}

	return component.SuccessResponse(component.DataMap{
		"count": count,
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

	return component.SuccessResponse("标记全部已读成功")
}

// DeleteNotificationReq 删除通知请求
type DeleteNotificationReq struct {
	NotificationId uint64 `json:"notificationId" validate:"required"`
}

// DeleteNotification 删除通知
func DeleteNotification(req component.BetterRequest[DeleteNotificationReq]) component.Response {
	err := eventNotification.DeleteNotification(req.Params.NotificationId, req.UserId)
	if err != nil {
		return component.FailResponse("删除通知失败")
	}

	return component.SuccessResponse("删除通知成功")
}

// GetNotificationTypesReq 获取通知类型请求
type GetNotificationTypesReq struct{}

// GetNotificationTypes 获取所有通知类型
func GetNotificationTypes(req component.BetterRequest[GetNotificationTypesReq]) component.Response {
	types := []component.DataMap{
		{"type": eventNotification.EventTypeComment, "name": "评论通知"},
		{"type": eventNotification.EventTypeReply, "name": "回复通知"},
		{"type": eventNotification.EventTypeSystem, "name": "系统通知"},
		{"type": eventNotification.EventTypeFollow, "name": "关注通知"},
	}

	return component.SuccessResponse(types)
}
