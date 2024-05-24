package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "admin:set",
		Short: "hexo tool",
		Run:   runAdminSet,
		//Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	cmd.Flags().Uint64("userId", 1, "input userId --userId=1")
	appendCommand(cmd)

}

func runAdminSet(cmd *cobra.Command, _ []string) {
	userId, err := cmd.Flags().GetUint64("userId")
	if err != nil {
		fmt.Println(err)
		return
	}
	// 判断有没有管理员角色
	roleEntity := role.Get(1)
	if roleEntity.Id == 0 {
		roleEntity.RoleName = "管理员"
		roleEntity.Effective = 1
		role.SaveOrCreateById(&roleEntity)
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

	ur := userRoleRs.GetByUserIdAndRoleId(userId, roleEntity.Id)
	if ur.Id == 0 {
		ur.RoleId = roleEntity.Id
		ur.UserId = userId
		ur.Effective = 1
		userRoleRs.SaveOrCreateById(&ur)
		fmt.Println("用户角色关系不存在，创建用户角色关系")
	}

}
