package migration

import (
	_ "embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/applySheet"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"log/slog"
)

func M() {
	// 数据库迁移
	migration(setting.UseMigration())
	dataInit()
}

func migration(migration bool) {
	if !migration {
		return
	}
	// 自动迁移
	var err error
	if !dbconnect.IsSqlite() {
		slog.Info("dbconnect 非sqlite不执行迁移")
		return
	} else {
		db := dbconnect.Connect()
		if err = db.AutoMigrate(
			&articleCategory.Entity{},
			&articleCategoryRs.Entity{},
			&articles.Entity{},
			&eventNotification.Entity{},
			&optRecord.Entity{},
			&pointsRecord.Entity{},
			&reply.Entity{},
			&role.Entity{},
			&rolePermissionRs.Entity{},
			&userFollow.Entity{},
			&userPoints.Entity{},
			&userRoleRs.Entity{},
			&users.Entity{},
			&taskQueue.Entity{},
			&articleLike.Entity{},
			&applySheet.Entity{},
			&pageConfig.Entity{},
		); err != nil {
			slog.Error("dbconnect migration err", "err", err)
		} else {
			slog.Info("dbconnect migration end")
		}
	}

	if !db4fileconnect.IsSqlite() {
		slog.Info("db4fileconnect 非sqlite不执行迁移")
		return
	} else {
		// 因为图片数据库比较大，所以单独迁移
		db4file := db4fileconnect.Connect()
		if err = db4file.AutoMigrate(
			&filedata.Entity{},
		); err != nil {
			slog.Error("db4fileconnect migration err", "err", err)
		} else {
			slog.Info("db4fileconnect migration end")
		}
	}
}

func dataInit() {
	user, _ := users.Get(1)
	if user.Id == 1 {
		return
	}
	adminUser := users.MakeUser("admin", "gooseforum", "admin@email.com")

	if err := users.Create(adminUser); err != nil {
		slog.Info("初始化用户失败")
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

	// 是否有用户
	// 是否有文章类别
	// 是否用文章
}

//go:embed initBlog.md
var initBlog string
