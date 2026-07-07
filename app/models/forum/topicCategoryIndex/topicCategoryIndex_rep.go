package topicCategoryIndex

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return builder().Create(entity).RowsAffected
	}

	return builder().Save(entity).RowsAffected
}

func GetByTopicId(topicId uint64) (entities []*Entity) {
	builder().Where("topic_id = ?", topicId).Find(&entities)
	return
}
