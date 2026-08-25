package fileUsage

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func CreateIfAbsent(entity *Entity) error {
	return builder().Clauses(clause.OnConflict{DoNothing: true}).Create(entity).Error
}

func GetByFileName(fileName string) ([]Entity, error) {
	if fileName == "" {
		return []Entity{}, nil
	}
	var entities []Entity
	err := builder().Where(queryopt.Eq("file_name", fileName)).Find(&entities).Error
	return entities, err
}

func ReplaceTargetUsages(targetType string, targetId uint64, usageTypes []string, usages []Entity) error {
	if len(usageTypes) == 0 {
		return nil
	}
	return dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		db := tx.Table(tableName)
		var existing []Entity
		if err := db.
			Where(queryopt.Eq("target_type", targetType)).
			Where(queryopt.Eq("target_id", targetId)).
			Where(queryopt.In("usage_type", usageTypes)).
			Find(&existing).Error; err != nil {
			return err
		}

		newFileNames := make([]string, 0, len(usages))
		for _, usage := range usages {
			if usage.FileName != "" {
				newFileNames = append(newFileNames, usage.FileName)
			}
		}
		var pending []Entity
		if len(newFileNames) > 0 {
			if err := tx.Table(tableName).
				Where(queryopt.Eq("target_type", TargetPendingUpload)).
				Where(queryopt.In("file_name", newFileNames)).
				Find(&pending).Error; err != nil {
				return err
			}
		}
		owners := uploadOwners(append(existing, pending...))
		if len(owners) > 0 {
			if err := tx.Table(tableName).Clauses(clause.OnConflict{DoNothing: true}).Create(&owners).Error; err != nil {
				return err
			}
		}
		if len(pending) > 0 {
			fileNames := make([]string, 0, len(pending))
			for _, usage := range pending {
				fileNames = append(fileNames, usage.FileName)
			}
			if err := tx.Table(tableName).
				Where(queryopt.Eq("target_type", TargetPendingUpload)).
				Where(queryopt.In("file_name", fileNames)).
				Delete(&Entity{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Table(tableName).
			Where(queryopt.Eq("target_type", targetType)).
			Where(queryopt.Eq("target_id", targetId)).
			Where(queryopt.In("usage_type", usageTypes)).
			Delete(&Entity{}).Error; err != nil {
			return err
		}
		if len(usages) == 0 {
			return nil
		}
		return tx.Table(tableName).Create(&usages).Error
	})
}

func uploadOwners(usages []Entity) []Entity {
	owners := make([]Entity, 0, len(usages))
	type ownerKey struct {
		fileName string
		userID   uint64
	}
	seen := make(map[ownerKey]struct{}, len(usages))
	for _, usage := range usages {
		if usage.FileName == "" || usage.UserId == 0 {
			continue
		}
		key := ownerKey{fileName: usage.FileName, userID: usage.UserId}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		owners = append(owners, Entity{
			FileName:   usage.FileName,
			TargetType: TargetUploadOwner,
			TargetId:   usage.UserId,
			UsageType:  UsageUploadOwner,
			UserId:     usage.UserId,
		})
	}
	return owners
}
