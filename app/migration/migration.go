package migration

import (
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/docs/docContents"
	"github.com/leancodebox/GooseForum/app/models/docs/docOperationLogs"
	"github.com/leancodebox/GooseForum/app/models/docs/docProjects"
	"github.com/leancodebox/GooseForum/app/models/docs/docVersions"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/applySheet"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCollection"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/kvstore"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
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
		&userOAuth.Entity{},
		&users.EntityComplete{},
		&taskQueue.Entity{},
		&articleLike.Entity{},
		&articleBookmark.Entity{},
		&applySheet.Entity{},
		&pageConfig.Entity{},
		&userStatistics.Entity{},
		&articleCollection.Entity{},
		&kvstore.Entity{},
		&docProjects.Entity{},
		&docVersions.Entity{},
		&docContents.Entity{},
		&docOperationLogs.Entity{},
	); err != nil {
		slog.Error("dbconnect migration err", "err", err)
	} else {
		slog.Info("dbconnect migration end")
	}

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

func initData() {
	category := articleCategory.GetOne()
	if category.Id == 0 {
		category.Category = "GooseForum"
		category.Desc = "🦢 大鹅栖息地 | 自由漫谈的江湖茶馆"
		articleCategory.SaveOrCreateById(&category)
		fmt.Println("标签不存在，创建标签")
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
