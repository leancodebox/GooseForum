package fileUsage

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func GetByFileName(fileName string) ([]Entity, error) {
	if fileName == "" {
		return []Entity{}, nil
	}
	var entities []Entity
	err := builder().Where(queryopt.Eq("file_name", fileName)).Find(&entities).Error
	return entities, err
}

func DeletePendingByFileNames(userID uint64, fileNames []string) error {
	if userID == 0 || len(fileNames) == 0 {
		return nil
	}
	return builder().
		Where(queryopt.Eq("target_type", TargetPendingUpload)).
		Where(queryopt.Eq("user_id", userID)).
		Where(queryopt.In("file_name", fileNames)).
		Delete(&Entity{}).Error
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
