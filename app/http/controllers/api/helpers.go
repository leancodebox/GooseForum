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
	configEntity := pageConfig.GetByPageType(pageType)
	configEntity.PageType = pageType
	configEntity.Config = jsonopt.Encode(config)
	pageConfig.CreateOrSave(&configEntity)
	clearCache()
	return component.SuccessResponse("success")
}
