package migration

import (
	_ "embed"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/chat/imConversations"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/chat/messages"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/migrationMapping"
	"github.com/leancodebox/GooseForum/app/models/forum/moderationLog"
	"github.com/leancodebox/GooseForum/app/models/forum/moderators"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/reports"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func M() {
	if !setting.UseMigration() {
		return
	}
	migrateSchema()
	runVersionedDataMigrations()
}

func migrateSchema() {
	var err error

	db := dbconnect.Connect()
	if err = db.AutoMigrate(
		&badges.Entity{},
		&eventNotification.Entity{},
		&fileUsage.Entity{},
		&moderationLog.Entity{},
		&migrationMapping.Entity{},
		&moderators.Entity{},
		&optRecord.Entity{},
		&pageConfig.Entity{},
		&pointsRecord.Entity{},
		&reports.Entity{},
		&topics.Entity{},
		&posts.Entity{},
		&category.Entity{},
		&topicCategoryIndex.Entity{},
		&topicUserAction.Entity{},
		&topicUserStat.Entity{},
		&role.Entity{},
		&rolePermissionRs.Entity{},
		&taskQueue.Entity{},
		&userFollow.Entity{},
		&userBadges.Entity{},
		&userOAuth.Entity{},
		&userPoints.Entity{},
		&users.EntityComplete{},
		&userStatistics.Entity{},
		&imConversations.Entity{},
		&imUserChatConfigs.Entity{},
		&messages.Entity{},
		&dailyStats.Entity{},
		&userActivities.Entity{},
	); err != nil {
		slog.Error("dbconnect migration err", "err", err)
	} else {
		slog.Info("dbconnect migration end")
	}

	db4file := db4fileconnect.Connect()
	if err = db4file.AutoMigrate(
		&filedata.Entity{},
	); err != nil {
		slog.Error("db4fileconnect migration err", "err", err)
	} else {
		slog.Info("db4fileconnect migration end")
	}

}
