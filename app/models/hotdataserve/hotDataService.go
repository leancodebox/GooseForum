package hotdataserve

import (
	"context"
	"errors"
	"github.com/allegro/bigcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/datacache"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"time"
)

var cacheResp *bigcache.BigCache

func init() {
	cacheResp, _ = bigcache.New(context.Background(), bigcache.DefaultConfig(1*time.Minute))
}

func GetOrLoad[T any](key string, load func() (T, error)) T {
	if cacheResp != nil {
		if data, err := cacheResp.Get(key); err == nil {
			return jsonopt.Decode[T](data)
		}
	}
	res, err := load()
	if cacheResp != nil && err == nil {
		cacheResp.Set(key, []byte(jsonopt.Encode(res)))
	}
	return res
}

func Reload[T any](key string, dataObj T) error {
	if cacheResp != nil {
		return cacheResp.Set(key, []byte(jsonopt.Encode(dataObj)))
	}
	return errors.New("no cache")
}

var footerConfigCache = &datacache.Cache[pageConfig.FooterConfig]{}

func GetFooterConfigCache() pageConfig.FooterConfig {
	data, _ := footerConfigCache.GetOrLoadE("", func() (pageConfig.FooterConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.FooterLinks, defaultconfig.GetDefaultFooter()), nil
	}, time.Minute)
	return data
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
	}, time.Minute)
	return data
}

var mailSettingsConfigCache = &datacache.Cache[pageConfig.MailSettingsConfig]{}

func GetMailSettingsConfigCache() pageConfig.MailSettingsConfig {
	data, _ := mailSettingsConfigCache.GetOrLoadE("", func() (pageConfig.MailSettingsConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.EmailSettings, pageConfig.MailSettingsConfig{}), nil
	}, time.Second*5)
	return data
}
