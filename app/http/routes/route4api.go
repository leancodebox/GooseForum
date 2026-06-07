package routes

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/api"
	"github.com/leancodebox/GooseForum/app/http/controllers/forum"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/resource"
)

func gzipEnabled() bool {
	return preferences.GetBool("server.gzip", true)
}

func assertRouter(ginApp *gin.Engine) {
	assetsFs, _ := resource.GetAssetsFS()
	staticFS, _ := resource.GetStaticFS()
	staticRoute := ginApp.Group("/")
	if gzipEnabled() {
		staticRoute.Use(middleware.CacheMiddleware)
		staticRoute.Use(gzip.Gzip(gzip.DefaultCompression))
		slog.Info("static assets gzip enabled", "cache", true)
	} else {
		slog.Info("static assets gzip disabled", "cache", false)
	}
	staticRoute.
		Use(middleware.BrowserCache).
		StaticFS("assets", http.FS(assetsFs)).
		StaticFS("static", http.FS(staticFS))
}

func viewRoute(ginApp *gin.Engine) {
	ginApp.GET("/reload", func(c *gin.Context) {
		if setting.IsProduction() {
			c.String(http.StatusNotFound, "404")
			return
		}
		forum.ReloadTemplates()
		c.String(200, "模板已刷新")
	})

	viewRouteApp := ginApp.Group("")
	viewRouteApp.Use(middleware.JWTAuth)
	if gzipEnabled() {
		viewRouteApp.Use(gzip.Gzip(gzip.DefaultCompression))
		slog.Info("view gzip enabled")
	} else {
		slog.Info("view gzip disabled")
	}

	viewRouteApp.GET("/", forum.Home)
	viewRouteApp.GET("/p/post/:id", forum.ArticleDetail)
	viewRouteApp.GET("/u/:userId", forum.UserProfile)
	viewRouteApp.GET("/c/:slug/:id", forum.Category)
	viewRouteApp.GET("/c/:slug/:id/l/:sort", forum.Category)
	viewRouteApp.GET("/links", forum.Links)
	viewRouteApp.GET("/sponsors", forum.Sponsors)
	viewRouteApp.GET("/messages", middleware.CheckLogin, forum.Messages)
	viewRouteApp.GET("/drafts", middleware.CheckLogin, forum.Drafts)
	viewRouteApp.GET("/settings", middleware.CheckLogin, forum.Settings)
	viewRouteApp.GET("/notifications", middleware.CheckLogin, forum.Notifications)
	viewRouteApp.GET("/publish", middleware.CheckLogin, forum.Publish)
	viewRouteApp.GET("/search", forum.Search)
	viewRouteApp.GET("/admin", middleware.CheckLogin, middleware.CheckAnyPermissionOrNotFound, forum.Manage)
	viewRouteApp.GET("/admin/*path", middleware.CheckLogin, middleware.CheckAnyPermissionOrNotFound, forum.Manage)
	viewRouteApp.GET("/login", forum.Login)
	viewRouteApp.GET("/reset-password", forum.ResetPassword)

	viewRouteApp.GET("/activate", controllers.ActivateAccount)
}

func siteInfoRoute(ginApp *gin.Engine) {
	ginApp.GET("/robots.txt", controllers.RenderRobotsTxt)
	ginApp.GET("/sitemap.xml", controllers.RenderSitemapXml)
	ginApp.GET("/rss.xml", controllers.RenderRss)
}

func apiRoute(ginApp *gin.Engine) {
	baseApi := ginApp.Group("api")

	baseApi.POST("login", controllers.Login)
	baseApi.GET("login-public-key", controllers.LoginPublicKey)
	baseApi.POST("register", controllers.Register)
	baseApi.POST("logout", controllers.Logout)

	baseApi.GET("get-captcha", UpQueryReq(api.GetCaptcha))
	baseApi.GET("user-card", UpQueryReq(api.GetUserCard))
	baseApi.GET("user-hover-card", UpQueryReq(api.GetUserHoverCard))
	baseApi.POST("forgot-password", UpButterReq(api.ForgotPassword))
	baseApi.POST("reset-password", UpButterReq(api.ResetPassword))
	baseApi.GET("auth/:provider", controllers.ProviderLogin)
	baseApi.GET("auth/:provider/callback", middleware.JWTAuth, controllers.ProviderCallback)

	loginApi := ginApp.Group("api").Use(middleware.JWTAuthCheck)
	loginApi.POST("set-user-info", UpButterReq(api.EditUserInfo))
	loginApi.POST("set-user-profile-cover", UpButterReq(api.EditUserProfileCover))
	loginApi.POST("set-user-email", UpButterReq(api.EditUserEmail))
	loginApi.POST("resend-activation-email", UpButterReq(api.ResendActivationEmail))
	loginApi.POST("set-user-name", UpButterReq(api.EditUsername))
	loginApi.POST("upload-avatar", api.UploadAvatar)
	loginApi.POST("change-password", UpButterReq(api.ChangePassword))
	loginApi.POST("auth/:provider/unbind", UpButterReq(controllers.UnbindOAuth))
	loginApi.GET("oauth/bindings", UpButterReq(controllers.GetOAuthBindings))

	forumApi := baseApi.Group("forum")
	forumApi.GET("get-site-statistics", ginUpNP(controllers.GetSiteStatistics))
	forumApi.GET("article-replies-window", middleware.JWTAuth, middleware.NoUpdateUserActivity, UpQueryReq(forum.ArticleRepliesWindow))

	forumLoginApi := forumApi.Use(middleware.JWTAuthCheck)
	forumLoginApi.GET("unread-status", middleware.NoUpdateUserActivity, UpButterReq(api.GetUnreadStatus))
	forumLoginApi.GET("notifications", middleware.NoUpdateUserActivity, UpQueryReq(api.NotificationList))
	forumLoginApi.POST("notification/mark-read", UpButterReq(api.MarkAsRead))
	forumLoginApi.POST("notification/mark-all-read", UpButterReq(api.MarkAllAsRead))
	forumLoginApi.POST("write-articles", UpButterReq(controllers.WriteArticles))
	forumLoginApi.POST("article-status", UpButterReq(controllers.UpdateArticleStatus))
	forumLoginApi.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	forumLoginApi.POST("articles-reply-update", UpButterReq(controllers.UpdateReply))
	forumLoginApi.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	forumLoginApi.POST("like-articles", UpButterReq(controllers.LikeArticle))
	forumLoginApi.POST("bookmark-article", UpButterReq(controllers.BookmarkArticle))
	forumLoginApi.POST("watch-article", UpButterReq(controllers.WatchArticle))
	forumLoginApi.POST("follow-user", UpButterReq(controllers.FollowUser))

	chatApi := forumApi.Group("chat", middleware.JWTAuthCheck)
	chatApi.POST("send", UpButterReq(api.SendMessage))
	chatApi.POST("messages", UpButterReq(api.GetMessages))
	chatApi.POST("mark-read", UpButterReq(api.MarkChatRead))

	adminApi := baseApi.Group("admin", middleware.JWTAuthCheck)

	adminApi.POST("traffic-overview", middleware.CheckPermission(permission.Admin), UpButterReq(api.GetTrafficOverview))

	adminApi.
		Group("", middleware.CheckPermission(permission.UserManager)).
		POST("user-list", UpButterReq(api.UserList)).
		POST("user-edit", UpButterReq(api.EditUser)).
		POST("user-badge-options", UpButterReq(api.UserBadgeOptions)).
		POST("save-user-badges", UpButterReq(api.SaveUserBadges)).
		GET("get-all-role-item", UpButterReq(api.GetAllRoleItem))

	adminApi.Group("", middleware.CheckPermission(permission.ArticlesManager)).
		POST("articles-list", UpButterReq(api.ArticlesList)).
		POST("article-source", UpButterReq(api.ArticleSource)).
		POST("article-edit", UpButterReq(api.EditArticle)).
		POST("article-delete", UpButterReq(api.DeleteArticle)).
		POST("article-pin-edit", UpButterReq(api.EditArticlePin)).
		POST("article-categories-edit", UpButterReq(api.EditArticleCategories)).
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

	adminApi.Group("", middleware.CheckPermission(permission.PageManager)).
		GET("friend-links", UpButterReq(api.GetFriendLinks)).
		POST("save-friend-links", UpButterReq(api.SaveFriendLinks)).
		GET("sponsors", UpButterReq(api.GetSponsors)).
		POST("save-sponsors", UpButterReq(api.SaveSponsors)).
		GET("announcement", UpButterReq(api.GetAnnouncement)).
		POST("save-announcement", UpButterReq(api.SaveAnnouncement))

	adminApi.Group("", middleware.CheckPermission(permission.SiteManager)).
		GET("server-version", UpButterReq(api.ServerVersion)).
		GET("site-settings", UpButterReq(api.GetSiteSettings)).
		POST("save-site-settings", UpButterReq(api.SaveSiteSettings)).
		GET("mail-settings", UpButterReq(api.GetMailSettings)).
		POST("save-mail-settings", UpButterReq(api.SaveMailSettings)).
		POST("test-mail-connection", UpButterReq(api.TestMailConnection)).
		GET("security-settings", UpButterReq(api.GetSecuritySettings)).
		POST("save-security-settings", UpButterReq(api.SaveSecuritySettings)).
		GET("posting-settings", UpButterReq(api.GetPostingSettings)).
		POST("save-posting-settings", UpButterReq(api.SavePostingSettings)).
		GET("badges", UpButterReq(api.BadgeList)).
		POST("badge-save", UpButterReq(api.SaveBadge)).
		POST("badge-delete", UpButterReq(api.DeleteBadge))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	r.POST("/img-upload", middleware.JWTAuthCheck, api.SaveImgByGinContext)
	r.GET("/img/*filename", api.GetFileByFileName)
}
