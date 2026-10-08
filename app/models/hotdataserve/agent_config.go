package hotdataserve

import (
	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

var agentSettingsCache = &sharedcache.Cache[pageConfig.AgentSettingsConfig]{Name: "agent-settings-config", MaxEntries: cacheconfig.Current().PageConfig}

func GetAgentSettingsConfigCache() pageConfig.AgentSettingsConfig {
	return agentSettingsCache.GetOrLoad("", pageConfig.GetAgentSettings, configFastCacheTTL)
}

func ClearAgentSettingsConfigCache() { agentSettingsCache.Clear() }
