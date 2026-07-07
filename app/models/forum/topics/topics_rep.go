package topics

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return builder().Create(entity).RowsAffected
	}

	return builder().Save(entity).RowsAffected
}

func Get(id uint64) (entity Entity) {
	builder().First(&entity, id)
	return
}

func GetMapByIds(ids []uint64) map[uint64]SmallEntity {
	var list []SmallEntity
	if len(ids) == 0 {
		return map[uint64]SmallEntity{}
	}
	builder().Where("id in ?", ids).Find(&list)
	result := make(map[uint64]SmallEntity, len(list))
	for _, item := range list {
		result[item.Id] = item
	}
	return result
}
