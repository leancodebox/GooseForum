package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

const (
	configFastCacheTTL = 5 * time.Second
	configSlowCacheTTL = time.Minute
	configRareCacheTTL = time.Hour
)

var sponsorsConfigCache = &sharedcache.Cache[pageConfig.SponsorsConfig]{Name: "sponsors-config", MaxEntries: cacheconfig.Current().PageConfig}

func SponsorsConfigCache() pageConfig.SponsorsConfig {
	return sponsorsConfigCache.GetOrLoad("", func() (pageConfig.SponsorsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SponsorsPage, defaultconfig.GetDefaultSponsorsConfig()), nil
	}, configSlowCacheTTL)
}

var siteSettingsConfigCache = &sharedcache.Cache[pageConfig.SiteSettingsConfig]{Name: "site-settings-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetSiteSettingsConfigCache() pageConfig.SiteSettingsConfig {
	return siteSettingsConfigCache.GetOrLoad("", func() (pageConfig.SiteSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteSettings, defaultconfig.GetDefaultSiteSettingsConfig()), nil
	}, configFastCacheTTL)
}

var siteThemeConfigCache = &sharedcache.Cache[pageConfig.SiteThemeConfig]{Name: "site-theme-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetSiteThemeConfigCache() pageConfig.SiteThemeConfig {
	return siteThemeConfigCache.GetOrLoad("", func() (pageConfig.SiteThemeConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteTheme, defaultconfig.GetDefaultSiteThemeConfig()), nil
	}, configFastCacheTTL)
}

var siteChromeConfigCache = &sharedcache.Cache[pageConfig.SiteChromeConfig]{Name: "site-chrome-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetSiteChromeConfigCache() pageConfig.SiteChromeConfig {
	return siteChromeConfigCache.GetOrLoad("", func() (pageConfig.SiteChromeConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteChrome, defaultconfig.GetDefaultSiteChromeConfig()), nil
	}, configFastCacheTTL)
}

var mailSettingsConfigCache = &sharedcache.Cache[pageConfig.MailSettingsConfig]{Name: "mail-settings-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetMailSettingsConfigCache() pageConfig.MailSettingsConfig {
	return mailSettingsConfigCache.GetOrLoad("", func() (pageConfig.MailSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.EmailSettings, defaultconfig.GetDefaultEmailSettingsConfig()), nil
	}, configFastCacheTTL)
}

var announcementConfigCache = &sharedcache.Cache[pageConfig.AnnouncementConfig]{Name: "announcement-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetAnnouncementConfigCache() pageConfig.AnnouncementConfig {
	return announcementConfigCache.GetOrLoad("", func() (pageConfig.AnnouncementConfig, error) {
		config := pageConfig.GetConfigByPageType(pageConfig.Announcement, defaultconfig.GetDefaultAnnouncementConfig())
		config.PrepareHTML()
		return config, nil
	}, configFastCacheTTL)
}

var securitySettingsConfigCache = &sharedcache.Cache[pageConfig.SecurityAndRegistration]{Name: "security-settings-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetSecuritySettingsConfigCache() pageConfig.SecurityAndRegistration {
	return securitySettingsConfigCache.GetOrLoad("", func() (pageConfig.SecurityAndRegistration, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SecuritySettings, defaultconfig.GetDefaultSecuritySettingsConfig()), nil
	}, configFastCacheTTL)
}

var postingSettingsConfigCache = &sharedcache.Cache[pageConfig.PostingContent]{Name: "posting-settings-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetPostingSettingsConfigCache() pageConfig.PostingContent {
	return postingSettingsConfigCache.GetOrLoad("", func() (pageConfig.PostingContent, error) {
		return pageConfig.GetPostingSettingsConfig(defaultconfig.GetDefaultPostingSettingsConfig()), nil
	}, configFastCacheTTL)
}

var httpNotifyConfigCache = &sharedcache.Cache[pageConfig.HttpNotifyConfig]{Name: "http-notify-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetHttpNotifyConfigCache() pageConfig.HttpNotifyConfig {
	return httpNotifyConfigCache.GetOrLoad("", func() (pageConfig.HttpNotifyConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.HttpNotify, defaultconfig.GetDefaultHttpNotifyConfig()), nil
	}, configRareCacheTTL)
}

var oauthSettingsConfigCache = &sharedcache.Cache[pageConfig.OAuthSettingsConfig]{Name: "oauth-settings-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetOAuthSettingsConfigCache() pageConfig.OAuthSettingsConfig {
	return oauthSettingsConfigCache.GetOrLoad("", func() (pageConfig.OAuthSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.OAuthSettings, defaultconfig.GetDefaultOAuthSettingsConfig()), nil
	}, configFastCacheTTL)
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

func ClearOAuthSettingsConfigCache() {
	oauthSettingsConfigCache.Clear()
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

var friendLinksConfigCache = &sharedcache.Cache[[]pageConfig.FriendLinksGroup]{Name: "friend-links-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetFriendLinksConfigCache() []pageConfig.FriendLinksGroup {
	return friendLinksConfigCache.GetOrLoad("", func() ([]pageConfig.FriendLinksGroup, error) {
		return pageConfig.GetConfigByPageType(pageConfig.FriendShipLinks, defaultconfig.GetDefaultFriendLinksConfig()), nil
	}, configSlowCacheTTL)
}

func ClearFriendLinksConfigCache() {
	friendLinksConfigCache.Clear()
}
