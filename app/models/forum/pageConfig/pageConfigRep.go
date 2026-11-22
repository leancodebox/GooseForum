package pageConfig

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetPostingSettingsConfig(defaultValue PostingContent) PostingContent {
	entity := GetByPageType(PostingSettings)
	if entity.Id == 0 {
		return defaultValue
	}

	config := jsonopt.Decode[PostingContent](entity.Config)
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

// SaveSiteConfig preserves avatar settings omitted by older clients. Compare-and-swap
// also protects a concurrent pool change between reading and writing the site settings.
func SaveSiteConfig(settings SiteSettingsConfig) error {
	for range 3 {
		var row Entity
		err := builder().Where("page_type = ?", SiteSettings).First(&row).Error
		next := settings
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = Entity{PageType: SiteSettings, Config: jsonopt.Encode(next)}
			result := builder().Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if result.Error != nil || result.RowsAffected == 1 {
				return result.Error
			}
			continue
		}
		if err != nil {
			return err
		}
		if next.AvatarPool == nil {
			previous, err := jsonopt.DecodeE[SiteSettingsConfig](row.Config)
			if err != nil {
				return err
			}
			if previous.AvatarPool != nil {
				next.AvatarPool = previous.AvatarPool
				next.DisableDefaultAvatar = previous.DisableDefaultAvatar
				next.DefaultAvatarFile = ""
			}
		}
		saved, err := CompareAndSwapConfig(row, jsonopt.Encode(next))
		if err != nil || saved {
			return err
		}
	}
	return errors.New("site settings changed concurrently")
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

const AppMigrationVersion uint32 = 31

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

func GetMigrationVersionWithDB(conn *gorm.DB) (uint32, error) {
	var entity Entity
	err := conn.Where("page_type = ?", Migration).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return cast.ToUint32(entity.Config), err
}

func SyncMigrationVersionWithDB(conn *gorm.DB, version uint32) error {
	entity := Entity{PageType: Migration, Config: cast.ToString(version)}
	return conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "page_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"config", "updated_at"}),
	}).Create(&entity).Error
}
