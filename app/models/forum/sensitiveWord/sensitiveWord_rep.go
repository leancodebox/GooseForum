package sensitiveWord

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

func builder() *gorm.DB { return dbconnect.Connect().Model(&Entity{}) }

func List(enabledOnly bool) []Entity {
	var list []Entity
	b := builder()
	if enabledOnly {
		b = b.Where("enabled = ?", true)
	}
	b.Order("id asc").Find(&list)
	return list
}

func Get(id uint64) (Entity, error) {
	var e Entity
	err := builder().First(&e, id).Error
	return e, err
}

func Save(e *Entity) error { return dbconnect.Connect().Save(e).Error }

func Delete(id uint64) error {
	result := builder().Delete(&Entity{}, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return result.Error
}

func LoadEnabled() ([]Entity, error) {
	var list []Entity
	err := builder().Where("enabled = ?", true).Order("id asc").Find(&list).Error
	return list, err
}
