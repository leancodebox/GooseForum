package api

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/agenttokens"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/mfaservice"
)

type CreateAgentTokenReq struct {
	Name     string   `json:"name"`
	Scopes   []string `json:"scopes"`
	Days     int      `json:"days"`
	Password string   `json:"password"`
	MFACode  string   `json:"mfaCode"`
}
type RevokeAgentTokenReq struct {
	ID string `json:"id"`
}

func ListAgentTokens(req component.BetterRequest[component.Null]) component.Response {
	if req.GinContext != nil {
		req.GinContext.Header("Cache-Control", "private, no-store")
	}
	rows := []agenttokens.Entity{}
	if err := dbconnect.Connect().Where("user_id = ?", req.UserId).Order("created_at DESC").Find(&rows).Error; err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(rows)
}

func CreateAgentToken(req component.BetterRequest[CreateAgentTokenReq]) component.Response {
	config := hotdataserve.GetAgentSettingsConfigCache()
	if !config.Enabled || !config.ManualTokens {
		return component.FailResponseCode(component.MessagePermissionDenied, nil)
	}
	user, err := req.GetUser()
	if err != nil {
		return component.FailResponseCode(component.MessageUserFetchFailed, nil)
	}
	if _, err := component.CheckUserPermission(&user, component.PermissionActionWrite); err != nil {
		return component.FailResponseError(err)
	}
	if user.Password == "" || algorithm.VerifyEncryptPassword(user.Password, req.Params.Password) != nil {
		return component.FailResponseCode(component.MessageAuthOldPasswordInvalid, nil)
	}
	if err := mfaservice.VerifySecondFactor(req.GinContext, user.Id, req.Params.MFACode); err != nil {
		return component.FailResponseCode(mfaErrorCode(err), nil)
	}
	if req.Params.Name == "" || len(req.Params.Name) > 100 || !agenttokens.ValidScopes(req.Params.Scopes) || req.Params.Days != 7 && req.Params.Days != 30 && req.Params.Days != 90 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	entity, raw, err := agenttokens.Create(dbconnect.Connect(), user.Id, user.TokenVersion, req.Params.Name, req.Params.Scopes, req.Params.Days)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	if req.GinContext != nil {
		req.GinContext.Header("Cache-Control", "no-store")
	}
	return component.SuccessResponse(map[string]any{"token": raw, "entry": entity})
}

func RevokeAgentToken(req component.BetterRequest[RevokeAgentTokenReq]) component.Response {
	if req.Params.ID == "" {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	err := dbconnect.Connect().Model(&agenttokens.Entity{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", req.Params.ID, req.UserId).Update("revoked_at", time.Now()).Error
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(true)
}

func RevokeAllAgentTokens(req component.BetterRequest[component.Null]) component.Response {
	err := dbconnect.Connect().Model(&agenttokens.Entity{}).Where("user_id = ? AND revoked_at IS NULL", req.UserId).Update("revoked_at", time.Now()).Error
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(true)
}

func DeleteAgentToken(req component.BetterRequest[RevokeAgentTokenReq]) component.Response {
	if req.Params.ID == "" {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	// Soft deletion invalidates authentication while retaining the source identity.
	err := dbconnect.Connect().Where("id = ? AND user_id = ?", req.Params.ID, req.UserId).Delete(&agenttokens.Entity{}).Error
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(true)
}
