package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/loginlogservice"
)

func ListAuthLogs(req component.BetterRequest[loginlogservice.Filter]) component.Response {
	if req.UserId == 0 {
		return component.FailResponseCode(component.MessageAuthRequired, nil)
	}
	page, err := loginlogservice.List(req.Params, req.UserId, false)
	if err != nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	return component.SuccessResponse(page)
}

// AdminAuthLogs is registered in the administrator-only route group.
func AdminAuthLogs(req component.BetterRequest[loginlogservice.Filter]) component.Response {
	page, err := loginlogservice.List(req.Params, 0, true)
	if err != nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	return component.SuccessResponse(page)
}
