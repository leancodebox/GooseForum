package messages

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return create(entity)
	} else {
		return save(entity)
	}
}

func Get(id any) (entity Entity) {
	builder().First(&entity, id)
	return
}

func GetByConvId(convId uint64, offset, limit int) []Entity {
	var entities []Entity
	builder().Where("conv_id = ?", convId).Order("created_at DESC").Offset(offset).Limit(limit).Find(&entities)
	return entities
}

func GetLatestByConvId(convId uint64, limit int) []Entity {
	var entities []Entity
	builder().Where("conv_id = ?", convId).Order("id DESC").Limit(limit).Find(&entities)
	return entities
}

func GetBeforeId(convId uint64, beforeId uint64, limit int) []Entity {
	var entities []Entity
	builder().Where("conv_id = ? AND id < ?", convId, beforeId).Order("id DESC").Limit(limit).Find(&entities)
	return entities
}

func GetAfterId(convId uint64, afterId uint64, limit int) []Entity {
	var entities []Entity
	builder().Where("conv_id = ? AND id > ?", convId, afterId).Order("id ASC").Limit(limit).Find(&entities)
	return entities
}

func MarkMessagesRead(convId, readerId uint64) {
	builder().Where("conv_id = ? AND sender_id != ? AND is_read = 0", convId, readerId).Update("is_read", 1)
}
