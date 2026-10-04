package role

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func GetForPolicy(id uint64) (Entity, error) {
	var entity Entity
	err := dbconnect.Connect().Where("id = ?", id).First(&entity).Error
	return entity, err
}

func ExistsEffective(id uint64) (bool, error) {
	var count int64
	err := dbconnect.Connect().Model(&Entity{}).Where("id = ? AND effective = ?", id, 1).Count(&count).Error
	return count > 0, err
}

func SaveForPolicy(entity *Entity) error {
	if entity.Id == 0 {
		return dbconnect.Connect().Create(entity).Error
	}
	return dbconnect.Connect().Model(&Entity{}).Where("id = ?", entity.Id).Update("role_name", entity.RoleName).Error
}

func DeleteForPolicy(entity *Entity) error { return dbconnect.Connect().Delete(entity).Error }
