package migration

import (
	"github.com/leancodebox/GooseForum/app/models/Users"
	"github.com/leancodebox/GooseForum/app/models/bbs/Articles"
	"github.com/leancodebox/GooseForum/app/models/bbs/Comment"
	"github.com/leancodebox/GooseForum/bundles/app"
	"github.com/leancodebox/GooseForum/bundles/dbconnect"
	"github.com/spf13/cast"
	"log/slog"
)

func M() {
	// 数据库迁移
	migration(app.UseMigration())
}

func migration(migration bool) {
	if migration == false {
		return
	}
	// 自动迁移
	var err error
	db := dbconnect.Connect()

	if err = db.AutoMigrate(
		&Users.Entity{},
		&Comment.Entity{},
		&Articles.Entity{},
	); err != nil {
		slog.Error(cast.ToString(err))
	} else {
		slog.Info("migration end")
	}
}
