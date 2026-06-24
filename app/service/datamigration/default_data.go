package datamigration

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func EnsureDefaultData() {
	category := articleCategory.GetOne()
	if category.Id == 0 {
		category.Category = "GooseForum"
		category.Desc = "🦢 大鹅栖息地 | 自由漫谈的江湖茶馆"
		articleCategory.SaveOrCreateById(&category)
		slog.Info("created default article category")
	}

	configEntity := pageConfig.GetByPageType(pageConfig.FriendShipLinks)
	if configEntity.Id == 0 {
		configEntity.PageType = pageConfig.FriendShipLinks
		configEntity.Config = jsonopt.Encode(defaultconfig.GetDefaultFriendLinksConfig())
		pageConfig.CreateOrSave(&configEntity)
	}
}
