package adminpolicyservice

import (
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

type Snapshot struct {
	AdminRoleIDs             []uint64
	EmailVerificationEnabled bool
}

// ReadSnapshot provides best-effort administrator checks without serializing management operations.
func ReadSnapshot() (Snapshot, error) {
	adminRoles, err := rolePermissionRs.RolesWithPermission(permission.Admin.Id())
	if err != nil {
		return Snapshot{}, err
	}
	settings, err := pageConfig.SecuritySettingsForPolicy(defaultconfig.GetDefaultSecuritySettingsConfig())
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{AdminRoleIDs: adminRoles, EmailVerificationEnabled: settings.EnableEmailVerification}, nil
}

func Permissions(roleIDs ...uint64) (map[uint64]map[uint64]bool, error) {
	grants, err := rolePermissionRs.PolicyGrants(roleIDs)
	if err != nil {
		return nil, err
	}
	permissions := map[uint64]map[uint64]bool{}
	for _, grant := range grants {
		if permissions[grant.RoleId] == nil {
			permissions[grant.RoleId] = map[uint64]bool{}
		}
		permissions[grant.RoleId][grant.PermissionId] = true
	}
	return permissions, nil
}
