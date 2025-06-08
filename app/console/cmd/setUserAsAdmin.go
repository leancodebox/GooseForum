package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "user:setAdmin",
		Short: "设置管理员",
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
	userEntity, _ := users.Get(userId)
	if userEntity.Id == 0 {
		fmt.Println("用户不存在")
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

	userEntity.RoleId = roleEntity.Id
	users.Save(&userEntity)

}
