package hotdataserve

import (
	"context"
	"errors"
	"github.com/allegro/bigcache/v3"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/datacache"
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

var footerConfigCache = &datacache.Cache[string, pageConfig.FooterConfig]{}

func GetFooterConfigCache() pageConfig.FooterConfig {
	data, _ := footerConfigCache.GetOrLoadE("", func() (pageConfig.FooterConfig, error) {
		return pageConfig.GetConfigByPageType(pageConfig.FooterLinks, defaultconfig.GetDefaultFooter()), nil
	}, time.Minute)
	return data
}
