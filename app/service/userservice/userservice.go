package userservice

import (
	_ "embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

//go:embed initBlog.md
var initBlog string

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

	ur := userRoleRs.GetByUserIdAndRoleId(adminUser.Id, roleEntity.Id)
	if ur.Id == 0 {
		ur.RoleId = roleEntity.Id
		ur.UserId = adminUser.Id
		ur.Effective = 1
		if err := userRoleRs.SaveOrCreateById(&ur); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("用户角色关系不存在，创建用户角色关系")
	}

	category := articleCategory.Get(1)
	if category.Id == 0 {
		category.Category = "GooseForum"
		articleCategory.SaveOrCreateById(&category)
		fmt.Println("标签不存在，创建标签")
	}
	controllers.WriteArticles(component.BetterRequest[controllers.WriteArticleReq]{
		Params: controllers.WriteArticleReq{
			Id:         0,
			Content:    initBlog,
			Title:      "Hi With GooseForum",
			Type:       1,
			CategoryId: []uint64{1},
		},
		UserId: adminUser.Id,
	})

}
