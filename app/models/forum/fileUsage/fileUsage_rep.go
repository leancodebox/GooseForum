package fileUsage

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func ReplaceTargetUsages(targetType string, targetId uint64, usageTypes []string, usages []Entity) error {
	if len(usageTypes) == 0 {
		return nil
	}
	db := builder()
	if err := db.
		Where(queryopt.Eq("target_type", targetType)).
		Where(queryopt.Eq("target_id", targetId)).
		Where(queryopt.In("usage_type", usageTypes)).
		Delete(&Entity{}).Error; err != nil {
		return err
	}
	if len(usages) == 0 {
		return nil
	}
	return db.Create(&usages).Error
}
