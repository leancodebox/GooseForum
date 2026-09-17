package api

import (
	"net/http"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/oauthservice"
)

func GetOAuthSettings(req component.BetterRequest[component.Null]) component.Response {
	return component.SuccessResponse(oauthservice.AdminSettings())
}

type SaveOAuthSettingsReq struct {
	Settings oauthservice.SettingsUpdate `json:"settings" validate:"required"`
}

func SaveOAuthSettings(req component.BetterRequest[SaveOAuthSettingsReq]) component.Response {
	if err := oauthservice.SaveSettings(req.Params.Settings); err != nil {
		return component.BuildResponse(http.StatusOK, component.ResultStruct{Code: component.FAIL, Message: err.Error()})
	}
	return component.SuccessResponse(oauthservice.AdminSettings())
}
