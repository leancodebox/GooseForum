package migration

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/Users"
	"github.com/leancodebox/GooseForum/app/models/bbs/Articles"
	"github.com/leancodebox/GooseForum/app/models/bbs/Comment"
	"github.com/leancodebox/GooseForum/bundles/app"
	"github.com/leancodebox/GooseForum/bundles/dbconnect"
	"github.com/leancodebox/GooseForum/bundles/logging"

	"gorm.io/gorm"
)

func M() {
	fmt.Println("init migration")
	// 数据库迁移
	migration(app.UseMigration(), dbconnect.Std())
}

func migration(migration bool, db *gorm.DB) {
	if migration == false {
		return
	}
	// 自动迁移
	var err error

	if err = db.AutoMigrate(
		&Users.Users{},
		&Comment.Comment{},
		&Articles.Articles{},
	); err != nil {
		logging.Error(err)
	} else {
		logging.Info("migration end")
	}
}
