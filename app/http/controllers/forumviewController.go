package controllers

import (
	_ "embed"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

type PageButton struct {
	Index int
	Page  int
}

func LoginView(c *gin.Context) {
	viewrender.SafeRender[any](c, "login.gohtml", nil, viewrender.NewPageMetaBuilder().
		SetTitle("登录/注册").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type ResetPasswordData struct {
	CommonDataVo
	PostDetailData
	LatestArticles      []*vo.ArticlesSimpleVo
	ArticleCategoryList []*articleCategory.Entity // 显式定义以解决嵌入结构体字段冲突
}

func ResetPasswordView(c *gin.Context) {
	viewrender.SafeRender[any](c, "reset_password.gohtml", ResetPasswordData{
		CommonDataVo: GetCommonData(c),
	}, viewrender.NewPageMetaBuilder().
		SetTitle("重置密码").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

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
