package accessGroupMembers

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
)

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
	if userID == 0 {
		return 0, nil
	}
	var count int64
	err := builder().
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
