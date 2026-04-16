package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/notificationservice"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func Home(c *gin.Context) {
	sort, _ := lo.Coalesce(c.Param("sort"), c.Query("sort"), "latest")

	page := lo.Ternary(cast.ToInt(c.Query("page")) <= 0, 1, cast.ToInt(c.Query("page")))

	// Use the paginated version
	latestArticles := hotdataserve.GetLatestArticlesSimpleVoPaginated(page, sort)

	nextPage := 0
	if len(latestArticles) == 20 {
		nextPage = page + 1
	}

	// If AJAX request, return JSON
	if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
		c.JSON(http.StatusOK, latestArticles)
		return
	}

	commonData := GetCommonData(c)
	active := "topics"
	switch sort {
	case "popular":
		active = "popular"
	case "hot":
		active = "hot"
	}
	commonData.Sidebar.SetActive(active)

	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle("GooseForum").
		SetDescription(commonData.GooseForumInfo.Desc).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()
	viewrender.SafeRender(c, "index.gohtml", HomeData{
		CommonDataVo:   commonData,
		LatestArticles: latestArticles,
		CurrentPage:    page,
		NextPage:       nextPage,
		Sort:           sort,
	}, pageMeta)
}

func Category(c *gin.Context) {
	idStr := c.Param("id")
	id := cast.ToUint64(idStr)
	category := hotdataserve.GetCategoryById(id)

	if category == nil {
		c.String(http.StatusNotFound, "Category not found")
		return
	}

	// Discourse 风格：如果 slug 不匹配，重定向到正确的 slug
	slug := c.Param("slug")
	sort := c.Param("sort")
	if slug != category.Category {
		redirectUrl := "/c/" + category.Category + "/" + idStr
		if sort != "" {
			redirectUrl += "/l/" + sort
		}
		c.Redirect(http.StatusMovedPermanently, redirectUrl)
		return
	}

	sort, _ = lo.Coalesce(sort, "latest")
	page := lo.Ternary(cast.ToInt(c.Query("page")) <= 0, 1, cast.ToInt(c.Query("page")))
	articlesData := hotdataserve.GetArticlesByCategorySimpleVo(id, sort, page)

	nextPage := 0
	if len(articlesData) == 20 {
		nextPage = page + 1
	}

	// 如果请求 JSON 格式（AJAX 加载更多），直接返回 JSON 数据
	if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
		c.JSON(http.StatusOK, articlesData)
		return
	}

	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle(category.Category + " - GooseForum").
		SetDescription(category.Desc).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()

	firstChar := ""
	if len(category.Category) > 0 {
		firstChar = string([]rune(category.Category)[0])
	}

	commonData := GetCommonData(c)
	commonData.Sidebar.SetActiveCategory(category.Id)

	viewrender.SafeRender(c, "category.gohtml", CategoryData{
		CommonDataVo:      commonData,
		Category:          category,
		CategoryFirstChar: firstChar,
		Articles:          articlesData,
		CurrentPage:       page,
		NextPage:          nextPage,
		Sort:              sort,
	}, pageMeta)
}

func User(c *gin.Context) {
	id := cast.ToUint64(c.Param("userId"))
	showUser := component.GetUserShowByUserId(id)
	if showUser.UserId == 0 {
		errorPage(c, "用户不存在", "用户不存在")
		return
	}

	last, _ := articles.GetLatestArticlesByUserId(id, 5)

	// 获取关注和粉丝列表（默认显示前10个）
	followingList := userFollow.GetFollowingList(id, 1, 10)
	followerList := userFollow.GetFollowerList(id, 1, 10)

	// 获取用户活动记录
	activities, _ := userActivities.GetUserTimeline(id, 0, 20)

	// 获取当前登录用户信息
	currentUserId := component.LoginUserId(c)

	// 检查当前用户是否关注了列表中的用户
	isFollowingAuthor := userFollow.IsFollowing(currentUserId, id)

	user, _ := users.Get(id)
	stats := userStatistics.Get(id)
	userCard := transform.User2UserCard(user, stats, isFollowingAuthor, currentUserId)
	myFollowingIds := userFollow.GetAllFollowingIds(currentUserId)

	viewrender.SafeRender(c, "user.gohtml", UserDataVo{
		CommonDataVo: GetCommonData(c),
		UserData: UserData{
			Articles:       hotdataserve.ArticlesSmallEntity2Vo(last),
			UserDetail:     userCard,
			FollowingList:  followingList,
			FollowerList:   followerList,
			Activities:     activities,
			MyFollowingIds: myFollowingIds,
			IsOwnProfile:   currentUserId == id,
		},
	}, viewrender.NewPageMetaBuilder().
		SetUserProfile(showUser.Username, showUser.Bio).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

func Messages(c *gin.Context) {
	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle("Messages - GooseForum").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()

	commonData := GetCommonData(c)
	commonData.Sidebar.SetActive("messages")
	viewrender.SafeRender(c, "chat.gohtml", MessagesData{
		CommonDataVo: commonData,
	}, pageMeta)
}

func Settings(c *gin.Context) {
	currentUserId := component.LoginUserId(c)
	// middleware checked login
	user, _ := users.Get(currentUserId)
	stats := userStatistics.Get(currentUserId)

	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle("Settings - GooseForum").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()

	viewrender.SafeRender(c, "settings.gohtml", SettingsData{
		CommonDataVo: GetCommonData(c),
		User:         transform.User2UserDetailedVo(user),
		Stats:        stats,
	}, pageMeta)
}

func NewTopic(c *gin.Context) {
	id := cast.ToUint64(c.Query("id"))
	pageTitle := "Create New Topic - GooseForum"
	if id > 0 {
		pageTitle = "Edit Topic - GooseForum"
	}

	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle(pageTitle).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()

	commonData := GetCommonData(c)
	commonData.Sidebar.SetActive("topics")
	viewrender.SafeRender(c, "publish.gohtml", NewTopicData{
		CommonDataVo: commonData,
		ArticleId:    id,
	}, pageMeta)
}

type NotificationsData struct {
	CommonDataVo
	Total   int64
	List    []*eventNotification.Entity
	HasMore bool
}

func Notifications(c *gin.Context) {
	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle("Notifications - GooseForum").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()
	userId := c.GetUint64("userId")
	total, notifications := notificationservice.GetNotificationItemList(userId, 20, 0, false)
	commonData := GetCommonData(c)
	commonData.Sidebar.SetActive("notifications")

	viewrender.SafeRender(c, "notifications.gohtml", NotificationsData{
		CommonDataVo: commonData,
		Total:        total,
		List:         notifications,
		HasMore:      len(notifications) >= 20,
	}, pageMeta)
}

type SponsorsData struct {
	CommonDataVo
	SponsorsInfo pageConfig.SponsorsConfig
}

func Sponsors(c *gin.Context) {
	sponsorsInfo := hotdataserve.SponsorsConfigCache()
	commonData := GetCommonData(c)
	commonData.Sidebar.SetActive("sponsors")
	viewrender.SafeRender(c, "sponsors.gohtml", SponsorsData{
		CommonDataVo: commonData,
		SponsorsInfo: sponsorsInfo,
	}, viewrender.NewPageMetaBuilder().
		SetTitle("Sponsors").
		SetDescription("GooseForum Sponsors").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type LinksData struct {
	CommonDataVo
	FriendLinksGroup []pageConfig.FriendLinksGroup
	TotalCounter     int
}

func Links(c *gin.Context) {
	res := hotdataserve.GetFriendLinksConfigCache()
	totalCounter := lo.SumBy(res, func(group pageConfig.FriendLinksGroup) int {
		return len(group.Links)
	})

	commonData := GetCommonData(c)
	commonData.Sidebar.SetActive("links")
	viewrender.SafeRender(c, "links.gohtml", LinksData{
		CommonDataVo:     commonData,
		FriendLinksGroup: res,
		TotalCounter:     totalCounter,
	}, viewrender.NewPageMetaBuilder().
		SetTitle("Friendly Links").
		SetDescription("GooseForum Friendly Links").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}
