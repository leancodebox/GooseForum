package migrationMapping

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return builder().Create(entity).RowsAffected
	}

	return builder().Save(entity).RowsAffected
}

func GetBySource(scope string, sourceType string, sourceId uint64) (entity Entity) {
	builder().Where("scope = ? AND source_type = ? AND source_id = ?", scope, sourceType, sourceId).First(&entity)
	return
}
