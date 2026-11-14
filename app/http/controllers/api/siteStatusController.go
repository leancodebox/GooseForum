package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/resource"
)

// GetSiteStatus keeps public deployment detection independent of user queries.
func GetSiteStatus(req component.BetterRequest[GetUnreadCountReq]) component.Response {
	req.GinContext.Header("Cache-Control", "private, no-store")
	result := component.DataMap{"deploymentId": resource.DeploymentID()}
	if req.UserId != 0 {
		unread := GetUnreadStatus(req)
		if unread.Data.Code == component.SUCCESS {
			result["unread"] = unread.Data.Result
		}
	}
	return component.SuccessResponse(result)
}
