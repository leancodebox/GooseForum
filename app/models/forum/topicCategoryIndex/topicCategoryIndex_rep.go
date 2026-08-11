package topicCategoryIndex

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/samber/lo"
	"gorm.io/gorm"
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
	return ReplaceTopicCategoriesWithDB(dbconnect.Connect(), topicId, categoryIDs)
}

func ReplaceTopicCategoriesWithDB(db *gorm.DB, topicId uint64, categoryIDs []uint64) error {
	categoryIDMap := lo.SliceToMap(categoryIDs, func(id uint64) (uint64, bool) {
		return id, true
	})
	var existing []*Entity
	if err := db.Where("topic_id = ?", topicId).Find(&existing).Error; err != nil {
		return err
	}
	for _, item := range existing {
		if _, ok := categoryIDMap[item.CategoryId]; ok {
			item.Effective = 1
			if err := db.Save(item).Error; err != nil {
				return err
			}
			delete(categoryIDMap, item.CategoryId)
			continue
		}
		item.Effective = 0
		if err := db.Save(item).Error; err != nil {
			return err
		}
	}
	for id := range categoryIDMap {
		rs := &Entity{TopicId: topicId, CategoryId: id, Effective: 1}
		if err := db.Create(rs).Error; err != nil {
			return err
		}
	}
	return nil
}
