package migration

import (
	_ "embed"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/chat/imConversations"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/chat/messages"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCollection"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/articlesUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/userUnreadCounts"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func M() {
	migration(setting.UseMigration())
	initData()
}

func migration(migration bool) {
	if !migration {
		return
	}
	var err error

	db := dbconnect.Connect()
	if err = db.AutoMigrate(
		&articleBookmark.Entity{},
		&articleCategory.Entity{},
		&articleCategoryRs.Entity{},
		&articleCollection.Entity{},
		&articleLike.Entity{},
		&articles.Entity{},
		&articlesUserStat.Entity{},
		&eventNotification.Entity{},
		&optRecord.Entity{},
		&pageConfig.Entity{},
		&pointsRecord.Entity{},
		&reply.Entity{},
		&role.Entity{},
		&rolePermissionRs.Entity{},
		&taskQueue.Entity{},
		&userFollow.Entity{},
		&userOAuth.Entity{},
		&userPoints.Entity{},
		&users.EntityComplete{},
		&userStatistics.Entity{},
		&userUnreadCounts.Entity{},
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

func initData() {
	category := articleCategory.GetOne()
	if category.Id == 0 {
		category.Category = "GooseForum"
		category.Desc = "🦢 大鹅栖息地 | 自由漫谈的江湖茶馆"
		articleCategory.SaveOrCreateById(&category)
		slog.Info("created default article category")
	}

	lItem := pageConfig.LinkItem{
		Name:    "GooseForum",
		Desc:    "简单的社区构建软件 / Easy forum software for building friendly communities.",
		Url:     "https://gooseforum.online",
		LogoUrl: "/static/pic/default-avatar.webp",
		Status:  1,
	}
	res := []pageConfig.FriendLinksGroup{
		{
			Name:  "community",
			Links: []pageConfig.LinkItem{lItem},
		},
		{
			Name:  "blog",
			Links: []pageConfig.LinkItem{lItem},
		},
		{
			Name:  "tool",
			Links: []pageConfig.LinkItem{lItem},
		},
	}
	configEntity := pageConfig.GetByPageType(pageConfig.FriendShipLinks)
	if configEntity.Id == 0 {
		configEntity.PageType = pageConfig.FriendShipLinks
		configEntity.Config = jsonopt.Encode(res)
		pageConfig.CreateOrSave(&configEntity)
	}
}
