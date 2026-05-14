package routes

import (
	"net/http"

	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/api"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/resource"
)

func assertRouter(ginApp *gin.Engine) {
	assetsFs, _ := resource.GetAssetsFS()
	staticFS, _ := resource.GetStaticFS()
	adminFs, _ := resource.GetAdminFS()
	ginApp.Group("/").
		Use(middleware.CacheMiddleware).
		Use(gzip.Gzip(gzip.DefaultCompression)).
		Use(middleware.BrowserCache).
		StaticFS("assets", http.FS(assetsFs)).
		StaticFS("static", http.FS(staticFS)).
		StaticFS("admin", http.FS(adminFs))
}

func viewRoute(ginApp *gin.Engine) {
	ginApp.GET("/reload", func(c *gin.Context) {
		if setting.IsProduction() {
			c.String(http.StatusNotFound, "404")
			return
		}
		viewrender.Reload()
		c.String(200, "模板已刷新")
	})

	viewRouteApp := ginApp.Group("")
	viewRouteApp.Use(middleware.JWTAuth).
		Use(gzip.Gzip(gzip.DefaultCompression))

	viewRouteApp.GET("/", controllers.Home)
	viewRouteApp.GET("/login", controllers.LoginView)
	viewRouteApp.GET("/c/:slug/:id", controllers.Category)
	viewRouteApp.GET("/c/:slug/:id/l/:sort", controllers.Category)
	viewRouteApp.GET("/p/post/:id", controllers.PostDetail)
	viewRouteApp.GET("/u/:userId", controllers.User)
	viewRouteApp.GET("/messages", middleware.CheckLogin, controllers.Messages)
	viewRouteApp.GET("/settings", middleware.CheckLogin, controllers.Settings)
	viewRouteApp.GET("/publish", middleware.CheckLogin, controllers.NewTopic)
	viewRouteApp.GET("/reset-password", controllers.ResetPasswordView)
	viewRouteApp.GET("/notifications", middleware.CheckLogin, controllers.Notifications)
	viewRouteApp.GET("/links", controllers.Links)
	viewRouteApp.GET("/sponsors", controllers.Sponsors)
	viewRouteApp.GET("/search", controllers.SearchPage)

	viewRouteApp.GET("/activate", controllers.ActivateAccount)
}

func siteInfoRoute(ginApp *gin.Engine) {
	ginApp.GET("/robots.txt", controllers.RenderRobotsTxt)
	ginApp.GET("/sitemap.xml", controllers.RenderSitemapXml)
	ginApp.GET("/rss.xml", controllers.RenderRssV2)
}

func apiRoute(ginApp *gin.Engine) {
	baseApi := ginApp.Group("api")

	baseApi.POST("login", controllers.Login)
	baseApi.GET("login-public-key", controllers.LoginPublicKey)
	baseApi.POST("register", controllers.Register)
	baseApi.POST("logout", controllers.Logout)

	baseApi.GET("get-captcha", UpQueryReq(api.GetCaptcha))
	baseApi.GET("user-card", UpQueryReq(api.GetUserCard))
	baseApi.GET("user-activities", UpQueryReq(api.GetUserActivities))
	baseApi.POST("forgot-password", UpButterReq(api.ForgotPassword))
	baseApi.POST("reset-password", UpButterReq(api.ResetPassword))
	baseApi.GET("auth/:provider", controllers.ProviderLogin)
	baseApi.GET("auth/:provider/callback", middleware.JWTAuth, controllers.ProviderCallback)

	loginApi := ginApp.Group("api").Use(middleware.JWTAuthCheck)
	loginApi.GET("get-user-info", UpButterReq(api.UserInfo))
	loginApi.POST("set-user-info", UpButterReq(api.EditUserInfo))
	loginApi.POST("set-user-email", UpButterReq(api.EditUserEmail))
	loginApi.POST("set-user-name", UpButterReq(api.EditUsername))
	loginApi.POST("invitation", UpButterReq(api.Invitation))
	loginApi.POST("upload-avatar", api.UploadAvatar)
	loginApi.POST("change-password", UpButterReq(api.ChangePassword))
	loginApi.POST("auth/:provider/unbind", UpButterReq(controllers.UnbindOAuth))
	loginApi.GET("oauth/bindings", UpButterReq(controllers.GetOAuthBindings))

	forumApi := baseApi.Group("forum")
	forumApi.GET("get-site-statistics", ginUpNP(controllers.GetSiteStatistics))
	forumApi.GET("get-articles-enum", ginUpNP(controllers.GetArticlesEnum))
	forumApi.POST("search-articles", UpButterReq(controllers.SearchArticles))

	forumLoginApi := forumApi.Use(middleware.JWTAuthCheck)
	forumLoginApi.POST("notification/list", UpButterReq(api.GetNotificationList))
	forumLoginApi.POST("notification/query", UpButterReq(api.QueryNotificationList))
	forumLoginApi.GET("notification/unread-count", UpButterReq(api.GetUnreadCount))
	forumLoginApi.GET("notification/last-unread", middleware.NoUpdateUserActivity, UpButterReq(api.GetLastUnread))
	forumLoginApi.POST("notification/mark-read", UpButterReq(api.MarkAsRead))
	forumLoginApi.POST("notification/mark-all-read", UpButterReq(api.MarkAllAsRead))
	forumLoginApi.POST("notification/delete", UpButterReq(api.DeleteNotification))
	forumLoginApi.GET("notification/types", UpButterReq(api.GetNotificationTypes))
	forumLoginApi.POST("get-articles-origin", middleware.CheckLogin, UpButterReq(controllers.WriteArticlesOrigin))
	forumLoginApi.POST("write-articles", UpButterReq(controllers.WriteArticles))
	forumLoginApi.POST("article-delete", UpButterReq(controllers.DeleteArticle))
	forumLoginApi.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	forumLoginApi.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	forumLoginApi.POST("get-user-articles", UpButterReq(controllers.GetUserArticles))
	forumLoginApi.POST("get-user-bookmarked-articles", UpButterReq(controllers.GetUserBookmarkedArticles))
	forumLoginApi.POST("like-articles", UpButterReq(controllers.LikeArticle))
	forumLoginApi.POST("bookmark-article", UpButterReq(controllers.BookmarkArticle))
	forumLoginApi.POST("follow-user", UpButterReq(controllers.FollowUser))

	chatApi := forumApi.Group("chat", middleware.JWTAuthCheck)
	chatApi.POST("send", UpButterReq(api.SendMessage))
	chatApi.POST("list", UpButterReq(api.GetChatList))
	chatApi.POST("messages", UpButterReq(api.GetMessages))
	chatApi.POST("mark-read", UpButterReq(api.MarkChatRead))
	chatApi.POST("delete", UpButterReq(api.DeleteChat))
	chatApi.POST("suggested-users", UpButterReq(api.GetSuggestedUsers))

	adminApi := baseApi.Group("admin", middleware.JWTAuthCheck)

	adminApi.POST("traffic-overview", UpButterReq(api.GetTrafficOverview))

	adminApi.
		Group("", middleware.CheckPermission(permission.UserManager)).
		POST("user-list", UpButterReq(api.UserList)).
		POST("user-edit", UpButterReq(api.EditUser)).
		GET("get-all-role-item", UpButterReq(api.GetAllRoleItem))

	adminApi.Group("", middleware.CheckPermission(permission.ArticlesManager)).
		POST("articles-list", UpButterReq(api.ArticlesList)).
		POST("article-edit", UpButterReq(api.EditArticle)).
		POST("category-list", UpButterReq(api.GetCategoryList)).
		POST("category-save", UpButterReq(api.SaveCategory)).
		POST("category-delete", UpButterReq(api.DeleteCategory))

	adminApi.Group("", middleware.CheckPermission(permission.RoleManager)).
		POST("get-permission-list", UpButterReq(api.GetPermissionList)).
		POST("role-list", UpButterReq(api.RoleList)).
		POST("role-save", UpButterReq(api.RoleSave)).
		POST("role-delete", UpButterReq(api.RoleDel))

	adminApi.Group("", middleware.CheckPermission(permission.Admin)).
		POST("opt-record-page", UpButterReq(api.OptRecordPage))

	adminApi.Group("", middleware.CheckPermission(permission.SiteManager)).
		GET("friend-links", UpButterReq(api.GetFriendLinks)).
		POST("save-friend-links", UpButterReq(api.SaveFriendLinks)).
		GET("site-settings", UpButterReq(api.GetSiteSettings)).
		POST("save-site-settings", UpButterReq(api.SaveSiteSettings)).
		GET("mail-settings", UpButterReq(api.GetMailSettings)).
		POST("save-mail-settings", UpButterReq(api.SaveMailSettings)).
		POST("test-mail-connection", UpButterReq(api.TestMailConnection)).
		GET("security-settings", UpButterReq(api.GetSecuritySettings)).
		POST("save-security-settings", UpButterReq(api.SaveSecuritySettings)).
		GET("posting-settings", UpButterReq(api.GetPostingSettings)).
		POST("save-posting-settings", UpButterReq(api.SavePostingSettings)).
		GET("sponsors", UpButterReq(api.GetSponsors)).
		POST("save-sponsors", UpButterReq(api.SaveSponsors)).
		GET("announcement", UpButterReq(api.GetAnnouncement)).
		POST("save-announcement", UpButterReq(api.SaveAnnouncement))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	r.POST("/img-upload", middleware.JWTAuthCheck, api.SaveImgByGinContext)
	r.GET("/img/*filename", middleware.BrowserCache, api.GetFileByFileName)
}
