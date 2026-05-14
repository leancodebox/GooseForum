package userservice

import (
	_ "embed"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

//go:embed initBlog.md
var initBlog string

func GetInitBlog() string {
	return initBlog
}

func FirstUserInit(adminUser *users.EntityComplete) {
	if adminUser.Id != 1 {
		return
	}

	roleEntity := role.Get(1)
	if roleEntity.Id == 0 {
		roleEntity.RoleName = "管理员"
		roleEntity.Effective = 1
		if err := role.SaveOrCreateById(&roleEntity); err != nil {
			slog.Error("create admin role failed", "error", err)
			return
		}
		slog.Info("created missing admin role")
	}

	rp := rolePermissionRs.GetRsByRoleIdAndPermission(roleEntity.Id, permission.Admin.Id())
	if rp.Id == 0 {
		rp.RoleId = roleEntity.Id
		rp.PermissionId = permission.Admin.Id()
		rp.Effective = 1
		rolePermissionRs.SaveOrCreateById(&rp)
		slog.Info("created missing admin role permission relation")
	}

	adminUser.RoleId = roleEntity.Id
	users.Save(adminUser)
}
