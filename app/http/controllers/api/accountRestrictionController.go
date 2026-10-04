package api

import (
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/accountrestrictions"
)

type UserRestrictionHistoryReq struct {
	UserId   uint64 `form:"userId"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

func UserRestrictionHistory(req component.BetterRequest[UserRestrictionHistoryReq]) component.Response {
	if req.Params.UserId == 0 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	page, size := max(req.Params.Page, 1), pageutil.BoundPageSize(req.Params.PageSize)
	records, total, err := accountrestrictions.HistoryPage(req.Params.UserId, page, size)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessPage(records, page, size, total)
}
