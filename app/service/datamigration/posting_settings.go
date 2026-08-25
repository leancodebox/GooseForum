package datamigration

import (
	"encoding/json"

	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

type PostingSettingsMigrationResult struct {
	Updated    bool
	Skipped    bool
	Failed     int
	LastFailed string
}

func EnsurePostingSettingsTopicLimit() PostingSettingsMigrationResult {
	result := PostingSettingsMigrationResult{}
	entity := pageConfig.GetByPageType(pageConfig.PostingSettings)
	if entity.Id == 0 {
		result.Skipped = true
		return result
	}

	defaultLimit := defaultconfig.GetDefaultPostingSettingsConfig().TextControl.MaxDailyTopicsPerUser
	config, changed, err := addDefaultTopicLimit(entity.Config, defaultLimit)
	if err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
		return result
	}
	if !changed {
		return result
	}
	entity.Config = config
	if pageConfig.CreateOrSave(&entity) == 0 {
		result.Failed = 1
		result.LastFailed = "save posting settings returned no affected rows"
		return result
	}
	result.Updated = true
	return result
}

func addDefaultTopicLimit(config string, defaultLimit int) (string, bool, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(config), &root); err != nil {
		return "", false, err
	}

	textControl := map[string]json.RawMessage{}
	if raw, ok := root["textControl"]; ok && len(raw) > 0 {
		if err := json.Unmarshal(raw, &textControl); err != nil {
			return "", false, err
		}
	}
	if _, exists := textControl["maxDailyTopicsPerUser"]; exists {
		return config, false, nil
	}

	limit, err := json.Marshal(defaultLimit)
	if err != nil {
		return "", false, err
	}
	textControl["maxDailyTopicsPerUser"] = limit
	textControlJSON, err := json.Marshal(textControl)
	if err != nil {
		return "", false, err
	}
	root["textControl"] = textControlJSON
	updated, err := json.Marshal(root)
	if err != nil {
		return "", false, err
	}
	return string(updated), true, nil
}
