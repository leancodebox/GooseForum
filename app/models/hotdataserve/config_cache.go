package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

const (
	configFastCacheTTL = 5 * time.Second
	configSlowCacheTTL = time.Minute
	configRareCacheTTL = time.Hour
)

var sponsorsConfigCache = &localcache.Cache[pageConfig.SponsorsConfig]{MaxEntries: cacheconfig.Current().PageConfig}

func SponsorsConfigCache() pageConfig.SponsorsConfig {
	return sponsorsConfigCache.GetOrLoad("", func() (pageConfig.SponsorsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SponsorsPage, defaultconfig.GetDefaultSponsorsConfig()), nil
	}, configSlowCacheTTL)
}

var siteSettingsConfigCache = &localcache.Cache[pageConfig.SiteSettingsConfig]{MaxEntries: cacheconfig.Current().PageConfig}

func GetSiteSettingsConfigCache() pageConfig.SiteSettingsConfig {
	return siteSettingsConfigCache.GetOrLoad("", func() (pageConfig.SiteSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteSettings, defaultconfig.GetDefaultSiteSettingsConfig()), nil
	}, configFastCacheTTL)
}

var siteThemeConfigCache = &localcache.Cache[pageConfig.SiteThemeConfig]{MaxEntries: cacheconfig.Current().PageConfig}

func GetSiteThemeConfigCache() pageConfig.SiteThemeConfig {
	return siteThemeConfigCache.GetOrLoad("", func() (pageConfig.SiteThemeConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteTheme, defaultconfig.GetDefaultSiteThemeConfig()), nil
	}, configFastCacheTTL)
}

var siteChromeConfigCache = &localcache.Cache[pageConfig.SiteChromeConfig]{MaxEntries: cacheconfig.Current().PageConfig}

func GetSiteChromeConfigCache() pageConfig.SiteChromeConfig {
	return siteChromeConfigCache.GetOrLoad("", func() (pageConfig.SiteChromeConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteChrome, defaultconfig.GetDefaultSiteChromeConfig()), nil
	}, configFastCacheTTL)
}

var mailSettingsConfigCache = &localcache.Cache[pageConfig.MailSettingsConfig]{MaxEntries: cacheconfig.Current().PageConfig}

func GetMailSettingsConfigCache() pageConfig.MailSettingsConfig {
	return mailSettingsConfigCache.GetOrLoad("", func() (pageConfig.MailSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.EmailSettings, defaultconfig.GetDefaultEmailSettingsConfig()), nil
	}, configFastCacheTTL)
}

var announcementConfigCache = &localcache.Cache[pageConfig.AnnouncementConfig]{MaxEntries: cacheconfig.Current().PageConfig}

func GetAnnouncementConfigCache() pageConfig.AnnouncementConfig {
	return announcementConfigCache.GetOrLoad("", func() (pageConfig.AnnouncementConfig, error) {
		config := pageConfig.GetConfigByPageType(pageConfig.Announcement, defaultconfig.GetDefaultAnnouncementConfig())
		config.PrepareHTML()
		return config, nil
	}, configFastCacheTTL)
}

var securitySettingsConfigCache = &localcache.Cache[pageConfig.SecurityAndRegistration]{MaxEntries: cacheconfig.Current().PageConfig}

func GetSecuritySettingsConfigCache() pageConfig.SecurityAndRegistration {
	return securitySettingsConfigCache.GetOrLoad("", func() (pageConfig.SecurityAndRegistration, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SecuritySettings, defaultconfig.GetDefaultSecuritySettingsConfig()), nil
	}, configFastCacheTTL)
}

var postingSettingsConfigCache = &localcache.Cache[pageConfig.PostingContent]{MaxEntries: cacheconfig.Current().PageConfig}

func GetPostingSettingsConfigCache() pageConfig.PostingContent {
	return postingSettingsConfigCache.GetOrLoad("", func() (pageConfig.PostingContent, error) {
		return pageConfig.GetConfigByPageType(pageConfig.PostingSettings, defaultconfig.GetDefaultPostingSettingsConfig()), nil
	}, configFastCacheTTL)
}

var httpNotifyConfigCache = &localcache.Cache[pageConfig.HttpNotifyConfig]{MaxEntries: cacheconfig.Current().PageConfig}

func GetHttpNotifyConfigCache() pageConfig.HttpNotifyConfig {
	return httpNotifyConfigCache.GetOrLoad("", func() (pageConfig.HttpNotifyConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.HttpNotify, defaultconfig.GetDefaultHttpNotifyConfig()), nil
	}, configRareCacheTTL)
}

func ClearSecuritySettingsConfigCache() {
	securitySettingsConfigCache.Clear()
}

func ClearPostingSettingsConfigCache() {
	postingSettingsConfigCache.Clear()
}

func ClearHttpNotifyConfigCache() {
	httpNotifyConfigCache.Clear()
}

func ClearSiteSettingsConfigCache() {
	siteSettingsConfigCache.Clear()
}

func ClearSiteThemeConfigCache() {
	siteThemeConfigCache.Clear()
}

func ClearSiteChromeConfigCache() {
	siteChromeConfigCache.Clear()
}

func ClearMailSettingsConfigCache() {
	mailSettingsConfigCache.Clear()
}

func ClearAnnouncementConfigCache() {
	announcementConfigCache.Clear()
}

func ClearSponsorsConfigCache() {
	sponsorsConfigCache.Clear()
}

var friendLinksConfigCache = &localcache.Cache[[]pageConfig.FriendLinksGroup]{MaxEntries: cacheconfig.Current().PageConfig}

func GetFriendLinksConfigCache() []pageConfig.FriendLinksGroup {
	return friendLinksConfigCache.GetOrLoad("", func() ([]pageConfig.FriendLinksGroup, error) {
		return pageConfig.GetConfigByPageType(pageConfig.FriendShipLinks, defaultconfig.GetDefaultFriendLinksConfig()), nil
	}, configSlowCacheTTL)
}

func ClearFriendLinksConfigCache() {
	friendLinksConfigCache.Clear()
}
