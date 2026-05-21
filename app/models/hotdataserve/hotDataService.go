package hotdataserve

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/datacache"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

var cacheResp *bigcache.BigCache

func init() {
	cacheResp, _ = bigcache.New(context.Background(), bigcache.DefaultConfig(1*time.Minute))
}

func GetOrLoad[T any](key string, load func() (T, error)) T {
	if cacheResp != nil {
		if data, err := cacheResp.Get(key); err == nil {
			slog.Debug("hotdataserve cache: hit", "key", key)
			return jsonopt.Decode[T](data)
		}
	}
	slog.Debug("hotdataserve cache: miss", "key", key)
	res, err := load()
	if err != nil {
		slog.Debug("hotdataserve cache: loader error", "key", key, "err", err)
		return res
	}
	if cacheResp != nil {
		if setErr := cacheResp.Set(key, []byte(jsonopt.Encode(res))); setErr != nil {
			slog.Debug("hotdataserve cache: store error", "key", key, "err", setErr)
		} else {
			slog.Debug("hotdataserve cache: stored", "key", key)
		}
	}
	return res
}

func Reload[T any](key string, dataObj T) error {
	if cacheResp != nil {
		if err := cacheResp.Set(key, []byte(jsonopt.Encode(dataObj))); err != nil {
			slog.Debug("hotdataserve cache: reload error", "key", key, "err", err)
			return err
		}
		slog.Debug("hotdataserve cache: reloaded", "key", key)
		return nil
	}
	return errors.New("no cache")
}

var sponsorsConfigCache = &datacache.Cache[pageConfig.SponsorsConfig]{}

func SponsorsConfigCache() pageConfig.SponsorsConfig {
	data, _ := sponsorsConfigCache.GetOrLoadE("", func() (pageConfig.SponsorsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SponsorsPage, pageConfig.SponsorsConfig{}), nil
	}, time.Minute)
	return data
}

var siteSettingsConfigCache = &datacache.Cache[pageConfig.SiteSettingsConfig]{}

func GetSiteSettingsConfigCache() pageConfig.SiteSettingsConfig {
	data, _ := siteSettingsConfigCache.GetOrLoadE("", func() (pageConfig.SiteSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SiteSettings, defaultconfig.GetDefaultSiteSettingsConfig()), nil
	}, time.Second*5)
	return data
}

var mailSettingsConfigCache = &datacache.Cache[pageConfig.MailSettingsConfig]{}

func GetMailSettingsConfigCache() pageConfig.MailSettingsConfig {
	data, _ := mailSettingsConfigCache.GetOrLoadE("", func() (pageConfig.MailSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.EmailSettings, defaultconfig.GetDefaultEmailSettingsConfig()), nil
	}, time.Second*5)
	return data
}

var announcementConfigCache = &datacache.Cache[pageConfig.AnnouncementConfig]{}

func GetAnnouncementConfigCache() pageConfig.AnnouncementConfig {
	data, _ := announcementConfigCache.GetOrLoadE("", func() (pageConfig.AnnouncementConfig, error) {
		config := pageConfig.GetConfigByPageType(pageConfig.Announcement, pageConfig.AnnouncementConfig{})
		config.PrepareHTML()
		return config, nil
	}, time.Second*5)
	return data
}

var securitySettingsConfigCache = &datacache.Cache[pageConfig.SecurityAndRegistration]{}

func GetSecuritySettingsConfigCache() pageConfig.SecurityAndRegistration {
	data, _ := securitySettingsConfigCache.GetOrLoadE("", func() (pageConfig.SecurityAndRegistration, error) {
		return pageConfig.GetConfigByPageType(pageConfig.SecuritySettings, defaultconfig.GetDefaultSecuritySettingsConfig()), nil
	}, time.Second*5)
	return data
}

var postingSettingsConfigCache = &datacache.Cache[pageConfig.PostingContent]{}

func GetPostingSettingsConfigCache() pageConfig.PostingContent {
	data, _ := postingSettingsConfigCache.GetOrLoadE("", func() (pageConfig.PostingContent, error) {
		return pageConfig.GetConfigByPageType(pageConfig.PostingSettings, defaultconfig.GetDefaultPostingSettingsConfig()), nil
	}, time.Second*5)
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

var friendLinksConfigCache = &datacache.Cache[[]pageConfig.FriendLinksGroup]{}

func GetFriendLinksConfigCache() []pageConfig.FriendLinksGroup {
	data, _ := friendLinksConfigCache.GetOrLoadE("", func() ([]pageConfig.FriendLinksGroup, error) {
		configEntity := pageConfig.GetByPageType(pageConfig.FriendShipLinks)
		res := jsonopt.Decode[[]pageConfig.FriendLinksGroup](configEntity.Config)
		return res, nil
	}, time.Minute)
	return data
}

func ClearFriendLinksConfigCache() {
	friendLinksConfigCache.Clear()
}
