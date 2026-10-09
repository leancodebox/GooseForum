package pageConfig

import (
	"encoding/json"
	"errors"
)

const AgentSettings = "agentSettings"

type AgentSettingsConfig struct {
	Enabled                bool `json:"enabled"`
	ManualTokens           bool `json:"manualTokens"`
	ReadPerMinute          int  `json:"readPerMinute"`
	UserReadPerMinute      int  `json:"userReadPerMinute"`
	WritePerMinute         int  `json:"writePerMinute"`
	AnonymousReadPerMinute int  `json:"anonymousReadPerMinute"`
	IPPerMinute            int  `json:"ipPerMinute"`
}

func DefaultAgentSettings() AgentSettingsConfig {
	return AgentSettingsConfig{Enabled: false, ManualTokens: false, ReadPerMinute: 60, UserReadPerMinute: 120, WritePerMinute: 10, AnonymousReadPerMinute: 30, IPPerMinute: 120}
}

func GetAgentSettings() (AgentSettingsConfig, error) {
	config := DefaultAgentSettings()
	var row struct{ Config string }
	result := builder().Model(&Entity{}).Where("page_type = ?", AgentSettings).Limit(1).Find(&row)
	if result.Error != nil {
		return AgentSettingsConfig{}, result.Error
	}
	if result.RowsAffected != 0 {
		if err := json.Unmarshal([]byte(row.Config), &config); err != nil {
			return AgentSettingsConfig{}, err
		}
	}
	if !config.Valid() {
		return AgentSettingsConfig{}, errors.New("invalid Agent settings")
	}
	return config, nil
}

func (config AgentSettingsConfig) Valid() bool {
	for _, value := range []int{config.ReadPerMinute, config.UserReadPerMinute, config.WritePerMinute, config.AnonymousReadPerMinute, config.IPPerMinute} {
		if value < 1 || value > 100000 {
			return false
		}
	}
	return true
}
