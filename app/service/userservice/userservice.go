package userservice

import (
	_ "embed"
	"fmt"
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

func FirstUserInit(adminUser *users.Entity) {
	if adminUser.Id != 1 {
		return
	}

	roleEntity := role.Get(1)
	if roleEntity.Id == 0 {
		roleEntity.RoleName = "管理员"
		roleEntity.Effective = 1
		if err := role.SaveOrCreateById(&roleEntity); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("角色不存在，创建角色")
	}

	rp := rolePermissionRs.GetRsByRoleIdAndPermission(roleEntity.Id, permission.Admin.Id())
	if rp.Id == 0 {
		rp.RoleId = roleEntity.Id
		rp.PermissionId = permission.Admin.Id()
		rp.Effective = 1
		rolePermissionRs.SaveOrCreateById(&rp)
		fmt.Println("角色权限关系不存在，创建角色权限关系")
	}

	adminUser.RoleId = roleEntity.Id
	users.Save(adminUser)
	
}
