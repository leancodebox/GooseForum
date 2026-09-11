package sensitivewordservice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"sync"
	"time"
)

var configCache struct {
	sync.Mutex
	value   pageConfig.SensitiveWordConfig
	expires time.Time
}

func ClearConfigCache() {
	configCache.Lock()
	configCache.expires = time.Time{}
	configCache.Unlock()
}

func Config() pageConfig.SensitiveWordConfig {
	configCache.Lock()
	defer configCache.Unlock()
	if time.Now().Before(configCache.expires) {
		return configCache.value
	}

	config := pageConfig.GetConfigByPageType(pageConfig.SensitiveWordSettings, pageConfig.SensitiveWordConfig{
		Enabled: false,
		Mode:    pageConfig.ModerationAfterReview,
	})
	if config.Mode != pageConfig.ModerationVisibleThenReview {
		config.Mode = pageConfig.ModerationAfterReview
	}
	configCache.value, configCache.expires = config, time.Now().Add(time.Minute)
	return config
}

func Enabled() bool { return Config().Enabled }
