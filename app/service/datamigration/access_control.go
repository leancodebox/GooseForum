package datamigration

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccessControlMigrationResult struct {
	Groups     int
	Categories int
	Grants     int
	Failed     int
	LastFailed string
}

func BackfillAccessControlDefaults() AccessControlMigrationResult {
	return BackfillAccessControlDefaultsWithDB(dbconnect.Connect())
}

func BackfillAccessControlDefaultsWithDB(conn *gorm.DB) AccessControlMigrationResult {
	result := AccessControlMigrationResult{}
	err := conn.Transaction(func(tx *gorm.DB) error {
		groups, err := ensureSystemAccessGroups(tx)
		if err != nil {
			return err
		}
		result.Groups = len(groups)

		var categories []category.Entity
		if err := tx.Order("id ASC").Find(&categories).Error; err != nil {
			return fmt.Errorf("load categories: %w", err)
		}
		result.Categories = len(categories)

		grants := make([]categoryGroupPermissions.Entity, 0, len(categories)*2)
		for _, item := range categories {
			grants = append(grants,
				categoryGroupPermissions.Entity{
					CategoryId:      item.Id,
					AccessGroupId:   groups[accessGroups.SystemKeyEveryone].Id,
					PermissionLevel: categoryGroupPermissions.PermissionRead,
					Status:          categoryGroupPermissions.StatusEnabled,
				},
				categoryGroupPermissions.Entity{
					CategoryId:      item.Id,
					AccessGroupId:   groups[accessGroups.SystemKeyRegistered].Id,
					PermissionLevel: categoryGroupPermissions.PermissionCreate,
					Status:          categoryGroupPermissions.StatusEnabled,
				},
			)
		}
		if len(grants) == 0 {
			return nil
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "category_id"}, {Name: "access_group_id"}},
			DoNothing: true,
		}).CreateInBatches(grants, 200).Error; err != nil {
			return fmt.Errorf("create compatibility grants: %w", err)
		}
		result.Grants = len(grants)
		return nil
	})
	if err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
	}
	return result
}

func ensureSystemAccessGroups(tx *gorm.DB) (map[string]accessGroups.Entity, error) {
	definitions := []accessGroups.Entity{
		{
			Name:      "Everyone",
			SystemKey: new(accessGroups.SystemKeyEveryone),
			JoinMode:  accessGroups.JoinModeSystem,
			Status:    accessGroups.StatusEnabled,
		},
		{
			Name:      "Registered",
			SystemKey: new(accessGroups.SystemKeyRegistered),
			JoinMode:  accessGroups.JoinModeSystem,
			Status:    accessGroups.StatusEnabled,
		},
	}
	for i := range definitions {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "system_key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "join_mode", "status", "updated_at",
			}),
		}).Create(&definitions[i]).Error; err != nil {
			return nil, fmt.Errorf("ensure system group %q: %w", *definitions[i].SystemKey, err)
		}
	}

	var stored []accessGroups.Entity
	if err := tx.Where("system_key IN ?", []string{
		accessGroups.SystemKeyEveryone,
		accessGroups.SystemKeyRegistered,
	}).Find(&stored).Error; err != nil {
		return nil, fmt.Errorf("reload system groups: %w", err)
	}
	result := make(map[string]accessGroups.Entity, len(stored))
	for _, group := range stored {
		if group.SystemKey != nil {
			result[*group.SystemKey] = group
		}
	}
	for _, key := range []string{accessGroups.SystemKeyEveryone, accessGroups.SystemKeyRegistered} {
		if result[key].Id == 0 {
			return nil, fmt.Errorf("system group %q missing after upsert", key)
		}
	}
	return result, nil
}

//go:fix inline
func stringPointer(value string) *string {
	return new(value)
}
