package api

import (
	"encoding/json"
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/trusthistory"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/trustlevelservice"
)

func GetTrustSettings(req component.BetterRequest[component.Null]) component.Response {
	c, err := pageConfig.GetTrustSettings()
	if err != nil {
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(c)
}

type SaveTrustSettingsReq struct {
	Settings json.RawMessage `json:"settings"`
}

func SaveTrustSettings(req component.BetterRequest[SaveTrustSettingsReq]) component.Response {
	settings, err := pageConfig.DecodeTrustSettings(req.Params.Settings)
	if err != nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	return savePageConfig(pageConfig.TrustSettings, settings, func() { hotdataserve.ClearTrustSettingsCache(); trustlevelservice.ClearReviewCache() })
}
func MyTrustLevel(req component.BetterRequest[component.Null]) component.Response {
	p, err := trustlevelservice.GetProgress(req.UserId)
	if err != nil {
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(p)
}

type TrustUserReq struct {
	UserId uint64 `json:"userId" form:"userId" validate:"required"`
	After  uint64 `json:"after" form:"after"`
}

func UserTrustLevel(req component.BetterRequest[TrustUserReq]) component.Response {
	p, err := trustlevelservice.GetProgress(req.Params.UserId)
	if err != nil {
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(p)
}
func UserTrustHistory(req component.BetterRequest[TrustUserReq]) component.Response {
	rows, err := trusthistory.List(req.Params.UserId, req.Params.After)
	if err != nil {
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(rows)
}

type SetTrustReq struct {
	UserId          uint64           `json:"userId" validate:"required"`
	Level           trustlevel.Level `json:"level"`
	Locked          bool             `json:"locked"`
	ExpectedVersion uint64           `json:"expectedVersion"`
	Reason          string           `json:"reason"`
}

func SetUserTrustLevel(req component.BetterRequest[SetTrustReq]) component.Response {
	err := trustlevelservice.SetLevel(req.UserId, req.Params.UserId, req.Params.Level, req.Params.Locked, req.Params.ExpectedVersion, req.Params.Reason)
	if err != nil {
		return component.FailResponseCode(component.MessageTrustAdjustment, component.MessageParams{"error": err.Error()})
	}
	p, err := trustlevelservice.GetProgress(req.Params.UserId)
	if err != nil {
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(p)
}

func trustWriteFailure(err error) (component.Response, bool) {
	var limit *trustlevelservice.LimitError
	if errors.As(err, &limit) {
		if limit.Kind == trustlevelservice.DailyTopicsKind {
			return component.FailResponseCode(component.MessageTrustDailyTopics, component.MessageParams{"limit": limit.Limit}), true
		}
		return component.FailResponseCode(component.MessageTrustContentLimit, component.MessageParams{"kind": limit.Kind, "limit": limit.Limit}), true
	}
	return component.Response{}, false
}
