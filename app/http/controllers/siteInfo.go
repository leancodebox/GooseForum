package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

type ForumInfo struct {
	Title        string
	Desc         string
	Independence bool
}

func GetGooseForumInfo() ForumInfo {
	siteData := hotdataserve.GetSiteSettingsConfigCache()
	return ForumInfo{
		Title:        siteData.SiteName,
		Desc:         siteData.SiteDescription,
		Independence: false,
	}
}

type UserData struct {
	Articles       []*vo.ArticlesSimpleVo
	UserDetail     *vo.UserCard
	FollowingList  []*users.EntityComplete
	FollowerList   []*users.EntityComplete
	Activities     []*userActivities.Entity
	MyFollowingIds []uint64
	IsOwnProfile   bool
}
