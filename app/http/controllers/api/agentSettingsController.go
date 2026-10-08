package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

type SaveAgentSettingsReq struct {
	Settings pageConfig.AgentSettingsConfig `json:"settings"`
}

func GetAgentSettings(req component.BetterRequest[component.Null]) component.Response {
	config, err := pageConfig.GetAgentSettings()
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(config)
}

func SaveAgentSettings(req component.BetterRequest[SaveAgentSettingsReq]) component.Response {
	if !req.Params.Settings.Valid() {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	return savePageConfig(pageConfig.AgentSettings, req.Params.Settings, hotdataserve.ClearAgentSettingsConfigCache)
}
