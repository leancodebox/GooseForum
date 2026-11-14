package topics

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"gorm.io/gorm"
)

// PublicMaxIDWithDB deliberately uses everyone grants, never a viewer's groups.
// A public auxiliary category cannot make a restricted main category readable.
func PublicMaxIDWithDB(db *gorm.DB) (uint64, error) {
	var id uint64
	err := db.Model(&Entity{}).
		Joins("JOIN category ON category.id = topics.main_category_id").
		Joins("JOIN category_group_permissions AS grants ON grants.category_id = topics.main_category_id").
		Joins("JOIN access_groups AS groups_public ON groups_public.id = grants.access_group_id").
		Where("topics.status = ? AND topics.process_status = ?", 1, 0).
		Where("groups_public.system_key = ? AND groups_public.status = ?", accessGroups.SystemKeyEveryone, accessGroups.StatusEnabled).
		Where("grants.status = ? AND grants.permission_level >= ?", categoryGroupPermissions.StatusEnabled, categoryGroupPermissions.PermissionRead).
		Select("COALESCE(MAX(topics.id), 0)").Scan(&id).Error
	return id, err
}
func PublicMaxID() (uint64, error) { return PublicMaxIDWithDB(dbconnect.Connect()) }
