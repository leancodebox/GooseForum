package userservice

import (
	_ "embed"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

//go:embed welcomeTopic.md
var welcomeTopic string

func GetWelcomeTopicContent() string {
	return welcomeTopic
}

func FirstUserInit(adminUser *users.EntityComplete) error {
	if adminUser.Id != 1 {
		return nil
	}

	roleEntity := role.Get(1)
	if roleEntity.Id == 0 {
		roleEntity.RoleName = "管理员"
		roleEntity.Effective = 1
		if err := role.SaveOrCreateById(&roleEntity); err != nil {
			slog.Error("create admin role failed", "error", err)
			return err
		}
		slog.Info("created missing admin role")
	}

	rp := rolePermissionRs.GetRsByRoleIdAndPermission(roleEntity.Id, permission.Admin.Id())
	if rp.Id == 0 {
		rp.RoleId = roleEntity.Id
		rp.PermissionId = permission.Admin.Id()
		rp.Effective = 1
		if err := rolePermissionRs.Save(&rp); err != nil {
			return err
		}
		permission.InvalidateRole(roleEntity.Id)
		slog.Info("created missing admin role permission relation")
	}

	adminUser.RoleId = roleEntity.Id
	if err := users.AssignInitialRole(adminUser.Id, roleEntity.Id); err != nil {
		slog.Error("save first admin user failed", "userId", adminUser.Id, "error", err)
		return err
	}
	RefreshUserCaches(adminUser)
	return nil
}
