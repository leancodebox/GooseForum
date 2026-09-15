package pageConfig

import (
	"encoding/json"
	"time"

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

// SaveConfig stores one typed page configuration and reports persistence
// errors to callers that need transactional configuration reloads.
func SaveConfig(pageType string, config string) error {
	entity := GetByPageType(pageType)
	entity.PageType = pageType
	entity.Config = config
	if entity.Id == 0 {
		return builder().Create(&entity).Error
	}
	return builder().Save(&entity).Error
}

// CompareAndSwapConfig updates one configuration only when it has not changed
// since the caller read it. This prevents runtime status updates from
// overwriting concurrent administrator changes.
func CompareAndSwapConfig(entity Entity, config string) (bool, error) {
	if entity.Id == 0 {
		return false, nil
	}
	result := builder().
		Where("id = ? AND updated_at = ? AND config = ?", entity.Id, entity.UpdatedAt, entity.Config).
		Updates(map[string]any{"config": config, "updated_at": time.Now()})
	return result.RowsAffected == 1, result.Error
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

const AppMigrationVersion uint32 = 23

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
