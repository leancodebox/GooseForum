package migration

import (
	"github.com/leancodebox/GooseForum/app/bundles/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cast"
	"log/slog"
)

func M() {
	// 数据库迁移
	migration(setting.UseMigration())
}

func migration(migration bool) {
	return
	if migration == false {
		return
	}
	// 自动迁移
	var err error
	db := dbconnect.Connect()

	if err = db.AutoMigrate(
		&users.Entity{},
		&articles.Entity{},
	); err != nil {
		slog.Error(cast.ToString(err))
	} else {
		slog.Info("migration end")
	}
}
