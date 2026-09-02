package pageConfig

import (
	"encoding/json"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/spf13/cast"
)

func GetPostingSettingsConfig(defaultValue PostingContent) PostingContent {
	entity := GetByPageType(PostingSettings)
	if entity.Id == 0 {
		return defaultValue
	}

	config := jsonopt.Decode[PostingContent](entity.Config)
	var raw struct {
		TextControl map[string]json.RawMessage `json:"textControl"`
	}
	if err := json.Unmarshal([]byte(entity.Config), &raw); err == nil {
		if _, exists := raw.TextControl["maxDailyTopicsPerUser"]; !exists {
			config.TextControl.MaxDailyTopicsPerUser = defaultValue.TextControl.MaxDailyTopicsPerUser
		}
	}
	return config
}

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func CreateOrSave(entity *Entity) int64 {
	if entity.Id == 0 {
		return create(entity)
	}

	return save(entity)
}

func GetByPageType(pageType string) (entity Entity) {
	builder().Where(queryopt.Eq(filedPageType, pageType)).First(&entity)
	return
}

func GetConfigByPageType[T any](pageType string, defaultValue T) T {
	var entity Entity
	builder().Where(queryopt.Eq(filedPageType, pageType)).First(&entity)
	if entity.Id > 0 {
		return jsonopt.Decode[T](entity.Config)
	}

	return defaultValue
}

const AppMigrationVersion uint32 = 19

func GetMigrationVersion() uint32 {
	configEntity := GetByPageType(Migration)
	return cast.ToUint32(configEntity.Config)
}

func SyncMigrationVersion(version uint32) error {
	configEntity := GetByPageType(Migration)
	configEntity.PageType = Migration
	configEntity.Config = cast.ToString(version)
	if configEntity.Id == 0 {
		return builder().Create(&configEntity).Error
	}
	return builder().Save(&configEntity).Error
}
