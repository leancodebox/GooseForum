package badges

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"gorm.io/gorm"
)

func All() (entities []*Entity) {
	builder().
		Order("sort_order ASC, id ASC").
		Find(&entities)
	return
}

func GetByCode(code string) (entity Entity) {
	builder().Where(queryopt.Eq("code", code)).First(&entity)
	return
}

func Save(entity *Entity) error {
	values := map[string]any{
		"code": entity.Code, "type": entity.Type, "grant_mode": entity.GrantMode,
		"name": entity.Name, "description": entity.Description,
		"icon_type": entity.IconType, "icon_key": entity.IconKey, "icon_url": entity.IconURL,
		"color": entity.Color, "level": entity.Level,
		"is_enabled": entity.IsEnabled, "is_wearable": entity.IsWearable,
		"sort_order": entity.SortOrder, "updated_at": time.Now(),
	}
	return dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		newCustom := entity.Id == 0 && entity.Type == TypeCustom
		if !newCustom {
			// A new system override only changes presentation; its statistics
			// must already have been initialized by the versioned migration.
			if _, err := badgeStats.LockWithDB(tx, entity.Code); err != nil {
				return err
			}
		}
		if entity.Id == 0 {
			// Map creation preserves explicit false values despite schema defaults.
			values["created_at"] = time.Now()
			if err := tx.Model(&Entity{}).Create(values).Error; err != nil {
				return err
			}
			if newCustom {
				return badgeStats.EnsureWithDB(tx, entity.Code)
			}
			return nil
		}
		return tx.Model(&Entity{}).Where("id = ?", entity.Id).Updates(values).Error
	})
}

func GetByCodeWithError(code string) (entity Entity, err error) {
	err = builder().Where(queryopt.Eq("code", code)).First(&entity).Error
	return
}

// DeleteUnusedCustom coordinates with grant writers through the same stats row.
func DeleteUnusedCustom(code string) (bool, error) {
	deleted := false
	err := dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		stat, err := badgeStats.LockWithDB(tx, code)
		if err != nil {
			return err
		}
		if stat.HasGrantHistory {
			return nil
		}
		result := tx.Where("code = ? AND type = ?", code, TypeCustom).Delete(&Entity{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		result = tx.Where("badge_code = ?", code).Delete(&badgeStats.Entity{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		deleted = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}
