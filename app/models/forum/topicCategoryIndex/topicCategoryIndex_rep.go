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

func ActiveCategoryIDsByTopicWithDB(db *gorm.DB, topicID uint64) ([]uint64, error) {
	if db == nil || topicID == 0 {
		return []uint64{}, nil
	}
	var categoryIDs []uint64
	err := db.Model(&Entity{}).
		Where("topic_id = ? AND effective = ?", topicID, 1).
		Order("category_id ASC").
		Pluck("category_id", &categoryIDs).Error
	return categoryIDs, err
}

func DeleteByTopicId(topicId uint64) int64 {
	return builder().Where("topic_id = ?", topicId).Delete(&Entity{}).RowsAffected
}

func DeleteByTopicIdWithDB(db *gorm.DB, topicID uint64) (int64, error) {
	result := db.Where("topic_id = ?", topicID).Delete(&Entity{})
	return result.RowsAffected, result.Error
}

func GetOneByCategoryId(categoryId uint64) (entity Entity) {
	builder().
		Where(queryopt.Eq("category_id", categoryId)).
		Where(queryopt.Eq("effective", 1)).
		First(&entity)
	return
}

func ActiveTopicIDsByCategoryWithDB(db *gorm.DB, categoryID uint64, maxIDs int) (topicIDs []uint64, complete bool, err error) {
	if db == nil || categoryID == 0 || maxIDs <= 0 {
		return []uint64{}, true, nil
	}
	err = db.Model(&Entity{}).
		Where("effective = ? AND category_id = ?", 1, categoryID).
		Order("topic_id ASC").
		Limit(maxIDs+1).
		Pluck("topic_id", &topicIDs).Error
	if err != nil {
		return nil, false, err
	}
	if len(topicIDs) > maxIDs {
		return topicIDs[:maxIDs], false, nil
	}
	return topicIDs, true, nil
}

func ActiveTopicIDsByCategory(categoryID uint64, maxIDs int) ([]uint64, bool, error) {
	return ActiveTopicIDsByCategoryWithDB(dbconnect.Connect(), categoryID, maxIDs)
}

type MultiCategoryTopicCount struct {
	CategoryID uint64 `gorm:"column:category_id"`
	TopicCount int64  `gorm:"column:topic_count"`
}

func MultiCategoryTopicCounts(categoryIDs []uint64) (map[uint64]int64, error) {
	return MultiCategoryTopicCountsWithDB(dbconnect.Connect(), categoryIDs)
}

func MultiCategoryTopicCountsWithDB(db *gorm.DB, categoryIDs []uint64) (map[uint64]int64, error) {
	counts := make(map[uint64]int64, len(categoryIDs))
	if db == nil || len(categoryIDs) == 0 {
		return counts, nil
	}
	var rows []MultiCategoryTopicCount
	err := db.Table(tableName+" AS current_idx").
		Select("current_idx.category_id AS category_id, COUNT(DISTINCT current_idx.topic_id) AS topic_count").
		Joins("JOIN "+tableName+" AS other_idx ON other_idx.topic_id = current_idx.topic_id AND other_idx.effective = ? AND other_idx.category_id <> current_idx.category_id", 1).
		Joins("JOIN topics ON topics.id = current_idx.topic_id AND topics.deleted_at IS NULL").
		Where("current_idx.effective = ? AND current_idx.category_id IN ?", 1, categoryIDs).
		Group("current_idx.category_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.CategoryID] = row.TopicCount
	}
	return counts, nil
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
