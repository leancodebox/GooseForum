package hotdataserve

import (
	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

var trustCache = &sharedcache.Cache[pageConfig.TrustConfig]{Name: "trust-settings", MaxEntries: 1}

func GetTrustSettings() (pageConfig.TrustConfig, error) {
	return trustCache.GetOrLoadE("", pageConfig.GetTrustSettings, configFastCacheTTL)
}
func ClearTrustSettingsCache() { trustCache.Clear() }
