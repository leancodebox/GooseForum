package hotdataserve

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

const (
	configFastCacheTTL = 5 * time.Second
	configSlowCacheTTL = time.Minute
	configCacheEntries = 4
)

var sponsorsConfigCache = &localcache.Cache[pageConfig.SponsorsConfig]{MaxEntries: configCacheEntries}

func SponsorsConfigCache() pageConfig.SponsorsConfig {
	data, _ := sponsorsConfigCache.GetOrLoadE("", func() (pageConfig.SponsorsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SponsorsPage, defaultconfig.GetDefaultSponsorsConfig()), nil
	}, configSlowCacheTTL)
	return data
}

var siteSettingsConfigCache = &localcache.Cache[pageConfig.SiteSettingsConfig]{MaxEntries: configCacheEntries}

func GetSiteSettingsConfigCache() pageConfig.SiteSettingsConfig {
	data, _ := siteSettingsConfigCache.GetOrLoadE("", func() (pageConfig.SiteSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteSettings, defaultconfig.GetDefaultSiteSettingsConfig()), nil
	}, configFastCacheTTL)
	return data
}

var mailSettingsConfigCache = &localcache.Cache[pageConfig.MailSettingsConfig]{MaxEntries: configCacheEntries}

func GetMailSettingsConfigCache() pageConfig.MailSettingsConfig {
	data, _ := mailSettingsConfigCache.GetOrLoadE("", func() (pageConfig.MailSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.EmailSettings, defaultconfig.GetDefaultEmailSettingsConfig()), nil
	}, configFastCacheTTL)
	return data
}

var announcementConfigCache = &localcache.Cache[pageConfig.AnnouncementConfig]{MaxEntries: configCacheEntries}

func GetAnnouncementConfigCache() pageConfig.AnnouncementConfig {
	data, _ := announcementConfigCache.GetOrLoadE("", func() (pageConfig.AnnouncementConfig, error) {
		config := pageConfig.GetConfigByPageType(pageConfig.Announcement, defaultconfig.GetDefaultAnnouncementConfig())
		config.PrepareHTML()
		return config, nil
	}, configFastCacheTTL)
	return data
}

var securitySettingsConfigCache = &localcache.Cache[pageConfig.SecurityAndRegistration]{MaxEntries: configCacheEntries}

func GetSecuritySettingsConfigCache() pageConfig.SecurityAndRegistration {
	data, _ := securitySettingsConfigCache.GetOrLoadE("", func() (pageConfig.SecurityAndRegistration, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SecuritySettings, defaultconfig.GetDefaultSecuritySettingsConfig()), nil
	}, configFastCacheTTL)
	return data
}

var postingSettingsConfigCache = &localcache.Cache[pageConfig.PostingContent]{MaxEntries: configCacheEntries}

func GetPostingSettingsConfigCache() pageConfig.PostingContent {
	data, _ := postingSettingsConfigCache.GetOrLoadE("", func() (pageConfig.PostingContent, error) {
		return pageConfig.GetConfigByPageType(pageConfig.PostingSettings, defaultconfig.GetDefaultPostingSettingsConfig()), nil
	}, configFastCacheTTL)
	return data
}

func ClearSecuritySettingsConfigCache() {
	securitySettingsConfigCache.Clear()
}

func ClearPostingSettingsConfigCache() {
	postingSettingsConfigCache.Clear()
}

func ClearSiteSettingsConfigCache() {
	siteSettingsConfigCache.Clear()
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

var friendLinksConfigCache = &localcache.Cache[[]pageConfig.FriendLinksGroup]{MaxEntries: configCacheEntries}

func GetFriendLinksConfigCache() []pageConfig.FriendLinksGroup {
	data, _ := friendLinksConfigCache.GetOrLoadE("", func() ([]pageConfig.FriendLinksGroup, error) {
		return pageConfig.GetConfigByPageType(pageConfig.FriendShipLinks, defaultconfig.GetDefaultFriendLinksConfig()), nil
	}, configSlowCacheTTL)
	return data
}

func ClearFriendLinksConfigCache() {
	friendLinksConfigCache.Clear()
}
