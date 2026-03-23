package userOAuth

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

// Create 创建OAuth记录
func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

// Update 更新OAuth记录
func Update(entity *Entity) error {
	return builder().Save(entity).Error
}

// Delete 删除OAuth记录
func Delete(id uint64) error {
	return builder().Delete(&Entity{}, id).Error
}

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return create(entity)
	}

	return save(entity)
}

// GetByProviderAndUID 根据提供商和UID获取OAuth记录
func GetByProviderAndUID(provider, providerUID string) *Entity {
	var entity Entity
	err := builder().Where("provider = ? AND provider_uid = ?", provider, providerUID).First(&entity).Error
	if err != nil {
		return nil
	}
	return &entity
}

// GetByUserIDAndProvider 根据用户ID和提供商获取OAuth记录
func GetByUserIDAndProvider(userID uint64, provider string) *Entity {
	var entity Entity
	err := builder().Where("user_id = ? AND provider = ?", userID, provider).First(&entity).Error
	if err != nil {
		return nil
	}
	return &entity
}

func Get(id any) (entity Entity) {
	builder().First(&entity, id)
	return
}

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

//func deleteEntity(entity *Entity) int64 {
//	result := builder().Delete(entity)
//	return result.RowsAffected
//}

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}
