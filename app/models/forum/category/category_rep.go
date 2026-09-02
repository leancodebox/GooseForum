package category

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm"
)

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

func Count() int64 {
	var count int64
	builder().Count(&count)
	return count
}

func All() (entities []*Entity) {
	builder().Order(queryopt.Asc("sort")).Order(queryopt.Asc("id")).Find(&entities)
	return
}

func DeleteEntity(entity *Entity) int64 {
	return builder().Delete(entity).RowsAffected
}

func AdjustTopicCountsWithDB(db *gorm.DB, incrementIDs, decrementIDs []uint64) error {
	if db == nil {
		return errors.New("category topic count database is required")
	}
	if len(incrementIDs) > 0 {
		result := db.Model(&Entity{}).
			Where("id IN ?", incrementIDs).
			UpdateColumn("topic_count", gorm.Expr("topic_count + 1"))
		if result.Error != nil {
			return result.Error
		}
	}
	if len(decrementIDs) > 0 {
		result := db.Model(&Entity{}).
			Where("id IN ?", decrementIDs).
			UpdateColumn("topic_count", gorm.Expr("CASE WHEN topic_count > 0 THEN topic_count - 1 ELSE 0 END"))
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
