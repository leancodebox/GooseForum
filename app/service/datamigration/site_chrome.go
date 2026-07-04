package datamigration

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

type legacySiteSettingsConfig struct {
	pageConfig.SiteSettingsConfig
	FooterInfo pageConfig.FooterInfo `json:"footerInfo"`
	BrandType  string                `json:"brandType"`
	BrandText  string                `json:"brandText"`
	BrandImage string                `json:"brandImage"`
}

type SiteChromeContentMigrationResult struct {
	Migrated bool
	Failed   int
}

func MigrateSiteChromeContent() SiteChromeContentMigrationResult {
	result := SiteChromeContentMigrationResult{}
	site, hasLegacySite := getLegacySiteSettingsConfig()
	chromeEntity := pageConfig.GetByPageType(pageConfig.SiteChrome)
	chrome := defaultconfig.GetDefaultSiteChromeConfig()
	if chromeEntity.Id > 0 {
		chrome = jsonopt.Decode[pageConfig.SiteChromeConfig](chromeEntity.Config)
	}

	if hasLegacySite {
		chrome.FooterInfo = cloneFooterInfo(site.FooterInfo)
		chrome.BrandType = normalizeBrandType(site.BrandType)
		chrome.BrandText = site.BrandText
		chrome.BrandImage = site.BrandImage
	}

	chromeEntity.PageType = pageConfig.SiteChrome
	chromeEntity.Config = jsonopt.Encode(chrome)
	if pageConfig.CreateOrSave(&chromeEntity) == 0 {
		result.Failed = 1
		slog.Error("site chrome content migration failed")
		return result
	}
	hotdataserve.ClearSiteChromeConfigCache()
	result.Migrated = true
	return result
}

func getLegacySiteSettingsConfig() (legacySiteSettingsConfig, bool) {
	entity := pageConfig.GetByPageType(pageConfig.SiteSettings)
	if entity.Id > 0 {
		return jsonopt.Decode[legacySiteSettingsConfig](entity.Config), true
	}
	return legacySiteSettingsConfig{}, false
}

func cloneFooterInfo(info pageConfig.FooterInfo) pageConfig.FooterInfo {
	return pageConfig.FooterInfo{
		Primary: append([]pageConfig.PItem(nil), info.Primary...),
		List:    append([]pageConfig.FooterItem(nil), info.List...),
	}
}

func normalizeBrandType(value string) string {
	switch value {
	case "text", "image":
		return value
	default:
		return "default"
	}
}
