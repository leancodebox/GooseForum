package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
)

type getNotificationReq struct {
}

func GetNotification(req component.BetterRequest[getNotificationReq]) component.Response {
	data := eventNotification.GetByUserId(req.UserId)
	return component.SuccessResponse(data)
}
