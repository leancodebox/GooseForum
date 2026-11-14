package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/loginlogservice"
)

func ListAuthLogs(req component.BetterRequest[loginlogservice.Filter]) component.Response {
	if req.UserId == 0 {
		return component.FailResponseCode(component.MessageAuthRequired, nil)
	}
	return authLogsResponse(req.Params, req.UserId, false)
}

// AdminAuthLogs is registered in the administrator-only route group.
func AdminAuthLogs(req component.BetterRequest[loginlogservice.Filter]) component.Response {
	return authLogsResponse(req.Params, 0, true)
}

func authLogsResponse(filter loginlogservice.Filter, userID uint64, admin bool) component.Response {
	page, err := loginlogservice.List(filter, userID, admin)
	if err != nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	return component.SuccessResponse(page)
}
