package topicCategoryIndex

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/samber/lo"
)

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

func DeleteByTopicId(topicId uint64) int64 {
	return builder().Where("topic_id = ?", topicId).Delete(&Entity{}).RowsAffected
}

func GetOneByCategoryId(categoryId uint64) (entity Entity) {
	builder().
		Where(queryopt.Eq("category_id", categoryId)).
		Where(queryopt.Eq("effective", 1)).
		First(&entity)
	return
}

func ReplaceTopicCategories(topicId uint64, categoryIDs []uint64) error {
	categoryIDMap := lo.SliceToMap(categoryIDs, func(id uint64) (uint64, bool) {
		return id, true
	})
	for _, item := range GetByTopicId(topicId) {
		if _, ok := categoryIDMap[item.CategoryId]; ok {
			item.Effective = 1
			if err := builder().Save(item).Error; err != nil {
				return err
			}
			delete(categoryIDMap, item.CategoryId)
			continue
		}
		item.Effective = 0
		if err := builder().Save(item).Error; err != nil {
			return err
		}
	}
	for id := range categoryIDMap {
		rs := &Entity{TopicId: topicId, CategoryId: id, Effective: 1}
		if err := builder().Create(rs).Error; err != nil {
			return err
		}
	}
	return nil
}
