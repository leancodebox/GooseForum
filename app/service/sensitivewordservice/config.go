package sensitivewordservice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func Config() pageConfig.SensitiveWordConfig {
	config := pageConfig.GetConfigByPageType(pageConfig.SensitiveWordSettings, pageConfig.SensitiveWordConfig{
		Enabled: false,
		Mode:    pageConfig.ModerationAfterReview,
	})
	if config.Mode != pageConfig.ModerationVisibleThenReview {
		config.Mode = pageConfig.ModerationAfterReview
	}
	return config
}

func Enabled() bool { return Config().Enabled }
