package rolePermissionRs

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

type PolicyGrant struct {
	RoleId       uint64
	PermissionId uint64
}

func PolicyGrants(roleIDs []uint64) ([]PolicyGrant, error) {
	var grants []PolicyGrant
	if len(roleIDs) == 0 {
		return grants, nil
	}
	err := dbconnect.Connect().Model(&Entity{}).Joins("JOIN role ON role.id = role_permission_rs.role_id AND role.effective = 1 AND role.deleted_at IS NULL").Where("role_permission_rs.role_id IN ? AND role_permission_rs.effective = ?", roleIDs, 1).Find(&grants).Error
	return grants, err
}

func RolesWithPermission(permissionID uint64) ([]uint64, error) {
	var grants []struct{ RoleId uint64 }
	err := dbconnect.Connect().Model(&Entity{}).Joins("JOIN role ON role.id = role_permission_rs.role_id AND role.effective = 1 AND role.deleted_at IS NULL").Where("role_permission_rs.permission_id = ? AND role_permission_rs.effective = ?", permissionID, 1).Find(&grants).Error
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(grants))
	for _, grant := range grants {
		ids = append(ids, grant.RoleId)
	}
	return ids, nil
}

func ReplacePolicyGrants(roleID uint64, permissions []uint64) error {
	if err := DeletePolicyGrants(roleID); err != nil {
		return err
	}
	if len(permissions) == 0 {
		return nil
	}
	grants := make([]Entity, 0, len(permissions))
	for _, id := range permissions {
		grants = append(grants, Entity{RoleId: roleID, PermissionId: id, Effective: 1})
	}
	return dbconnect.Connect().Create(&grants).Error
}

func DeletePolicyGrants(roleID uint64) error {
	return dbconnect.Connect().Where("role_id = ?", roleID).Delete(&Entity{}).Error
}
