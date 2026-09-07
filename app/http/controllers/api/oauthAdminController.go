package api

import (
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
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(oauthservice.AdminSettings())
}
