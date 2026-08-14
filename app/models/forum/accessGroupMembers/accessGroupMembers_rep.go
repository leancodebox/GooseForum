package accessGroupMembers

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"gorm.io/gorm"
)

const categoryGrantStatusEnabled int8 = 1

func ActiveGroupIDsByUser(userID uint64) ([]uint64, error) {
	if userID == 0 {
		return []uint64{}, nil
	}
	var groupIDs []uint64
	err := builder().
		Joins("JOIN access_groups ON access_groups.id = access_group_members.access_group_id AND access_groups.status = ? AND access_groups.system_key IS NULL", accessGroups.StatusEnabled).
		Where(queryopt.Eq("access_group_members.user_id", userID)).
		Where(queryopt.Eq("access_group_members.status", StatusEnabled)).
		Order(queryopt.Asc("access_group_members.access_group_id")).
		Pluck("access_group_members.access_group_id", &groupIDs).Error
	return groupIDs, err
}

// ActiveUserIDsWithCategoryCapability filters a notification/audience batch in
// one query. System groups are handled by accesscontrol before this call; this
// query only considers enabled custom-group memberships and enabled grants.
func ActiveUserIDsWithCategoryCapability(userIDs []uint64, categoryID uint64, minimumLevel int8) ([]uint64, error) {
	if len(userIDs) == 0 || categoryID == 0 {
		return []uint64{}, nil
	}
	var readableUserIDs []uint64
	err := builder().
		Distinct("access_group_members.user_id").
		Joins("JOIN access_groups ON access_groups.id = access_group_members.access_group_id AND access_groups.status = ? AND access_groups.system_key IS NULL", accessGroups.StatusEnabled).
		Joins("JOIN category_group_permissions ON category_group_permissions.access_group_id = access_group_members.access_group_id AND category_group_permissions.status = ?", categoryGrantStatusEnabled).
		Where(queryopt.In("access_group_members.user_id", userIDs)).
		Where(queryopt.Eq("access_group_members.status", StatusEnabled)).
		Where(queryopt.Eq("category_group_permissions.category_id", categoryID)).
		Where("category_group_permissions.permission_level >= ?", minimumLevel).
		Order(queryopt.Asc("access_group_members.user_id")).
		Pluck("access_group_members.user_id", &readableUserIDs).Error
	return readableUserIDs, err
}

func ManagedGroupIDsByUser(userID uint64) ([]uint64, error) {
	if userID == 0 {
		return []uint64{}, nil
	}
	var groupIDs []uint64
	err := builder().
		Joins("JOIN access_groups ON access_groups.id = access_group_members.access_group_id AND access_groups.status = ? AND access_groups.system_key IS NULL", accessGroups.StatusEnabled).
		Where(queryopt.Eq("access_group_members.user_id", userID)).
		Where(queryopt.Eq("access_group_members.member_role", MemberRoleManager)).
		Where(queryopt.Eq("access_group_members.status", StatusEnabled)).
		Order(queryopt.Asc("access_group_members.access_group_id")).
		Pluck("access_group_members.access_group_id", &groupIDs).Error
	return groupIDs, err
}

func CountActiveCustomGroupsByUser(userID uint64) (int64, error) {
	return CountActiveCustomGroupsByUserWithDB(builder(), userID)
}

func CountActiveCustomGroupsByUserWithDB(db *gorm.DB, userID uint64) (int64, error) {
	if userID == 0 {
		return 0, nil
	}
	var count int64
	err := db.Model(&Entity{}).
		Joins("JOIN access_groups ON access_groups.id = access_group_members.access_group_id AND access_groups.status = ? AND access_groups.system_key IS NULL", accessGroups.StatusEnabled).
		Where(queryopt.Eq("access_group_members.user_id", userID)).
		Where(queryopt.Eq("access_group_members.status", StatusEnabled)).
		Count(&count).Error
	return count, err
}

func ByGroupID(groupID uint64) ([]Entity, error) {
	var entities []Entity
	err := builder().Where(queryopt.Eq("access_group_id", groupID)).Order(queryopt.Asc("id")).Find(&entities).Error
	return entities, err
}

func All() ([]Entity, error) {
	var entities []Entity
	err := builder().Order(queryopt.Asc("access_group_id")).Order(queryopt.Asc("id")).Find(&entities).Error
	return entities, err
}

func GetByGroupUser(groupID, userID uint64) (Entity, error) {
	var entity Entity
	err := builder().Where(queryopt.Eq("access_group_id", groupID)).Where(queryopt.Eq("user_id", userID)).First(&entity).Error
	return entity, err
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

func Delete(entity *Entity) error {
	return builder().Delete(entity).Error
}
