package sensitivewordservice

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

var configCache = sharedcache.Cache[pageConfig.SensitiveWordConfig]{Name: "sensitive-word-config", MaxEntries: 1}

func ClearConfigCache() {
	configCache.Clear()
}

func Config() pageConfig.SensitiveWordConfig {
	return configCache.GetOrLoad("", loadConfig, time.Minute)
}

func loadConfig() (pageConfig.SensitiveWordConfig, error) {
	config := pageConfig.GetConfigByPageType(pageConfig.SensitiveWordSettings, pageConfig.SensitiveWordConfig{
		Enabled: false,
		Mode:    pageConfig.ModerationAfterReview,
	})
	if config.Mode != pageConfig.ModerationVisibleThenReview {
		config.Mode = pageConfig.ModerationAfterReview
	}
	return config, nil
}

func Enabled() bool { return Config().Enabled }
