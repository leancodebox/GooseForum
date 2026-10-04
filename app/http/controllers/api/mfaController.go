package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/loginlogservice"
	"github.com/leancodebox/GooseForum/app/service/mfaservice"
	"net/http"
)

func mfaErrorCode(err error) component.MessageCode {
	if errors.Is(err, mfaservice.ErrUnavailable) {
		return component.MessageMFAUnavailable
	}
	return component.MessageMFAInvalid
}

type MFARequest struct {
	Password string `json:"password"`
	Code     string `json:"code"`
	Action   string `json:"action"`
}

type AdminMFARequest struct {
	UserId uint64 `json:"userId"`
	Reason string `json:"reason"`
}

func AdminMFAStatus(req component.BetterRequest[AdminMFARequest]) component.Response {
	if req.Params.UserId == 0 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	status, err := mfaservice.Status(req.Params.UserId)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(status)
}

func AdminMFAReset(req component.BetterRequest[AdminMFARequest]) component.Response {
	if req.Params.UserId == 0 || len(req.Params.Reason) > 200 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if err := mfaservice.AdminResetBy(req.UserId, req.Params.UserId, req.Params.Reason); err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(nil)
}

func MFALogin(c *gin.Context) {
	var req MFARequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Code) > 128 {
		loginlogservice.Failure(c, "mfa", "", "challenge_input_invalid")
		c.JSON(http.StatusOK, component.FailDataCode(component.MessageRequestInvalidParams, nil))
		return
	}
	redirect, err := mfaservice.Login(c, req.Code)
	if err != nil {
		c.JSON(http.StatusOK, component.FailDataCode(mfaErrorCode(err), nil))
		return
	}
	c.JSON(http.StatusOK, component.SuccessData(map[string]any{"redirect": redirect}))
}
func MFAStatus(req component.BetterRequest[component.Null]) component.Response {
	data, err := mfaservice.Status(req.UserId)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(data)
}
func MFABegin(req component.BetterRequest[MFARequest]) component.Response {
	req.GinContext.Header("Cache-Control", "no-store")
	data, err := mfaservice.Begin(req.GinContext, req.UserId, req.Params.Password)
	if err != nil {
		loginlogservice.Record(req.GinContext, loginlogservice.Event{UserID: req.UserId, Action: "security_failure", Method: "mfa", Result: "failure", Reason: "setup_rejected"})
		return component.FailResponseCode(mfaErrorCode(err), nil)
	}
	return component.SuccessResponse(data)
}
func MFAChange(req component.BetterRequest[MFARequest]) component.Response {
	req.GinContext.Header("Cache-Control", "no-store")
	if len(req.Params.Code) > 128 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	codes, err := mfaservice.Change(req.GinContext, req.UserId, req.Params.Password, req.Params.Code, req.Params.Action)
	if err != nil {
		loginlogservice.Record(req.GinContext, loginlogservice.Event{UserID: req.UserId, Action: "security_failure", Method: "mfa", Result: "failure", Reason: "change_rejected"})
		return component.FailResponseCode(mfaErrorCode(err), nil)
	}
	return component.SuccessResponse(map[string]any{"recoveryCodes": codes})
}
