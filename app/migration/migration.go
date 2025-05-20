package migration

import (
	_ "embed"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/goose/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers"
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
	"log/slog"
)

func M() {
	// 数据库迁移
	migration(setting.UseMigration())
	initData()
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

func initData() {
	lItem := controllers.LinkItem{
		Name:    "GooseForum",
		Desc:    "简单的社区构建软件 / Easy forum software for building friendly communities.",
		Url:     "https://gooseforum.online",
		LogoUrl: "/static/pic/default-avatar.png",
	}
	res := []controllers.FriendLinksGroup{
		{
			Name:  "community",
			Links: []controllers.LinkItem{lItem},
		},
		{
			Name:  "blog",
			Links: []controllers.LinkItem{lItem},
		},
		{
			Name:  "tool",
			Links: []controllers.LinkItem{lItem},
		},
	}
	configEntity := pageConfig.GetByPageType(controllers.FriendShipLinks)
	if configEntity.Id == 0 {
		configEntity.PageType = controllers.FriendShipLinks
		configEntity.Config = jsonopt.Encode(res)
		pageConfig.CreateOrSave(&configEntity)
	}
}
