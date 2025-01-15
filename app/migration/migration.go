package migration

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleTag"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
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
}

func migration(migration bool) {
	// 自动迁移
	var err error
	if !dbconnect.IsSqlite() {
		slog.Info("非sqlite不执行迁移")
		return
	}
	db := dbconnect.Connect()
	if err = db.AutoMigrate(
		&articleCategory.Entity{},
		&articleCategoryRs.Entity{},
		&articles.Entity{},
		&articleTag.Entity{},
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
	); err != nil {
		slog.Error("migration err", "err", err)
	} else {
		slog.Info("migration end")
	}

	// 因为图片数据库比较大，所以单独迁移
	db4file := db4fileconnect.Connect()
	if err = db4file.AutoMigrate(
		&filedata.Entity{},
	); err != nil {
		slog.Error("migration err", "err", err)
	} else {
		slog.Info("migration end")
	}
}
