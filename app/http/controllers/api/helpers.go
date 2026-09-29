package api

import (
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func dataMap(key string, value any) component.DataMap {
	return component.DataMap{key: value}
}

func successDataMap(key string, value any) component.Response {
	return component.SuccessResponse(dataMap(key, value))
}

func savePageConfig(pageType string, config any, clearCache func()) component.Response {
	if err := pageConfig.SaveConfig(pageType, jsonopt.Encode(config)); err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	clearCache()
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}
