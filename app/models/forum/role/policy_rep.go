package role

func GetForPolicy(id uint64) (Entity, error) {
	var entity Entity
	err := builder().Where("id = ?", id).First(&entity).Error
	return entity, err
}

func ExistsEffective(id uint64) (bool, error) {
	var count int64
	err := builder().Model(&Entity{}).Where("id = ? AND effective = ?", id, 1).Count(&count).Error
	return count > 0, err
}

func SaveForPolicy(entity *Entity) error {
	if entity.Id == 0 {
		return builder().Create(entity).Error
	}
	return builder().Model(&Entity{}).Where("id = ?", entity.Id).Update("role_name", entity.RoleName).Error
}

func DeleteForPolicy(entity *Entity) error { return builder().Delete(entity).Error }
