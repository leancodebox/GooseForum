package routes

import (
	"net/http"
	"time"

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

	//
	viewRouteApp.GET("/", controllers.Home)
	viewRouteApp.GET("/login", controllers.LoginView)
	// 301 重定向旧路径到新路径
	viewRouteApp.GET("/post/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(http.StatusMovedPermanently, "/p/post/"+id)
	})
	viewRouteApp.GET("/user/:userId", func(c *gin.Context) {
		userId := c.Param("userId")
		c.Redirect(http.StatusMovedPermanently, "/u/"+userId)
	})
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

	// 添加激活路由
	viewRouteApp.GET("/activate", controllers.ActivateAccount)
}

func siteInfoRoute(ginApp *gin.Engine) {
	// SEO 相关路由
	ginApp.GET("/robots.txt", controllers.RenderRobotsTxt)
	ginApp.GET("/sitemap.xml", controllers.RenderSitemapXml)
	ginApp.GET("/rss.xml", controllers.RenderRssV2)
}

func apiRoute(ginApp *gin.Engine) {
	// 非登陆下的用户操作
	baseApi := ginApp.Group("api")

	baseApi.POST("login", controllers.Login)
	baseApi.POST("register", controllers.Register)
	baseApi.POST("logout", controllers.Logout)

	// 验证码
	baseApi.GET("get-captcha", UpQueryReq(api.GetCaptcha))
	// 用户卡片信息
	baseApi.GET("user-card", UpQueryReq(api.GetUserCard))
	// 用户活动记录
	baseApi.GET("user-activities", UpQueryReq(api.GetUserActivities))
	// 忘记密码和重置密码路由
	baseApi.POST("forgot-password", UpButterReq(api.ForgotPassword))
	baseApi.POST("reset-password", UpButterReq(api.ResetPassword))
	// GitHub OAuth 路由
	baseApi.GET("auth/:provider", controllers.ProviderLogin)
	baseApi.GET("auth/:provider/callback", middleware.JWTAuth, controllers.ProviderCallback)

	// 登陆状态下的用户操作
	loginApi := ginApp.Group("api").Use(middleware.JWTAuthCheck)
	// 用户信息
	loginApi.GET("get-user-info", UpButterReq(api.UserInfo))
	// 设置用户信息
	loginApi.POST("set-user-info", UpButterReq(api.EditUserInfo))
	loginApi.POST("set-user-email", UpButterReq(api.EditUserEmail))
	loginApi.POST("set-user-name", UpButterReq(api.EditUsername))
	// 邀请码
	loginApi.POST("invitation", UpButterReq(api.Invitation))
	// 上传头像
	loginApi.POST("upload-avatar", api.UploadAvatar)
	// 修改密码
	loginApi.POST("change-password", UpButterReq(api.ChangePassword))
	// OAuth 解绑和绑定状态查询
	loginApi.POST("auth/:provider/unbind", UpButterReq(controllers.UnbindOAuth))
	loginApi.GET("oauth/bindings", UpButterReq(controllers.GetOAuthBindings))

	forumApi := baseApi.Group("forum")
	forumApi.POST("apply-link-add", UpButterReq(api.ApplyAddLink))
	// 站点统计
	forumApi.GET("get-site-statistics", ginUpNP(controllers.GetSiteStatistics))
	// 分类列表
	forumApi.GET("get-articles-enum", ginUpNP(controllers.GetArticlesEnum))
	// 搜索文章
	forumApi.POST("search-articles", UpButterReq(controllers.SearchArticles))

	forumLoginApi := forumApi.Use(middleware.JWTAuthCheck)
	// 通知相关接口
	forumLoginApi.POST("notification/list", UpButterReq(api.GetNotificationList))
	forumLoginApi.POST("notification/query", UpButterReq(api.QueryNotificationList))
	forumLoginApi.GET("notification/unread-count", UpButterReq(api.GetUnreadCount))
	forumLoginApi.GET("notification/last-unread", middleware.NoUpdateUserActivity, UpButterReq(api.GetLastUnread))
	forumLoginApi.POST("notification/mark-read", UpButterReq(api.MarkAsRead))
	forumLoginApi.POST("notification/mark-all-read", UpButterReq(api.MarkAllAsRead))
	forumLoginApi.POST("notification/delete", UpButterReq(api.DeleteNotification))
	forumLoginApi.GET("notification/types", UpButterReq(api.GetNotificationTypes))
	// 编辑文章时原始文章内容
	forumLoginApi.POST("get-articles-origin", middleware.CheckLogin, UpButterReq(controllers.WriteArticlesOrigin))
	// 发布文章
	forumLoginApi.POST("write-articles", UpButterReq(controllers.WriteArticles))
	// 删除文章
	forumLoginApi.POST("article-delete", UpButterReq(controllers.DeleteArticle))
	// 回复文章
	forumLoginApi.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	// 回复评论
	forumLoginApi.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	// 用户文章列表
	forumLoginApi.POST("get-user-articles", UpButterReq(controllers.GetUserArticles))
	forumLoginApi.POST("get-user-bookmarked-articles", UpButterReq(controllers.GetUserBookmarkedArticles))
	// 文章点赞
	forumLoginApi.POST("like-articles", UpButterReq(controllers.LikeArticle))
	// 文章收藏
	forumLoginApi.POST("bookmark-article", UpButterReq(controllers.BookmarkArticle))
	// 关注
	forumLoginApi.POST("follow-user", UpButterReq(controllers.FollowUser))

	// 私信相关接口
	chatApi := forumApi.Group("chat", middleware.JWTAuthCheck)
	chatApi.POST("send", UpButterReq(api.SendMessage))
	chatApi.POST("list", UpButterReq(api.GetChatList))
	chatApi.POST("messages", UpButterReq(api.GetMessages))
	chatApi.POST("mark-read", UpButterReq(api.MarkChatRead))
	chatApi.POST("delete", UpButterReq(api.DeleteChat))
	chatApi.POST("suggested-users", UpButterReq(api.GetSuggestedUsers))

	adminApi := baseApi.Group("admin", middleware.JWTAuthCheck)

	adminApi.POST("traffic-overview", UpButterReq(api.GetTrafficOverview))

	// 用户管理
	adminApi.
		Group("", middleware.CheckPermission(permission.UserManager)).
		POST("user-list", UpButterReq(api.UserList)).
		POST("user-edit", UpButterReq(api.EditUser)).
		GET("get-all-role-item", UpButterReq(api.GetAllRoleItem))

	// 文章管理 &  分类管理
	adminApi.Group("", middleware.CheckPermission(permission.ArticlesManager)).
		POST("articles-list", UpButterReq(api.ArticlesList)).
		POST("article-edit", UpButterReq(api.EditArticle)).
		POST("category-list", UpButterReq(api.GetCategoryList)).
		POST("category-save", UpButterReq(api.SaveCategory)).
		POST("category-delete", UpButterReq(api.DeleteCategory))

	// 权限管理
	adminApi.Group("", middleware.CheckPermission(permission.RoleManager)).
		POST("get-permission-list", UpButterReq(api.GetPermissionList)).
		POST("role-list", UpButterReq(api.RoleList)).
		POST("role-save", UpButterReq(api.RoleSave)).
		POST("role-delete", UpButterReq(api.RoleDel))

	// 操作日志
	adminApi.Group("", middleware.CheckPermission(permission.Admin)).
		POST("opt-record-page", UpButterReq(api.OptRecordPage))

	// 站点管理
	adminApi.Group("", middleware.CheckPermission(permission.SiteManager)).
		POST("apply-sheet-list", UpButterReq(api.ApplySheet)).
		POST("apply-sheet-update", UpButterReq(api.UpdateApplySheet)).
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
		GET("footer-links", UpButterReq(api.GetFooterLinks)).
		POST("save-footer-links", UpButterReq(api.SaveFooterLinks)).
		GET("sponsors", UpButterReq(api.GetSponsors)).
		POST("save-sponsors", UpButterReq(api.SaveSponsors)).
		GET("announcement", UpButterReq(api.GetAnnouncement)).
		POST("save-announcement", UpButterReq(api.SaveAnnouncement))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	// 文件上传接口 - 每日最多上传n张图片
	r.POST("/img-upload", middleware.JWTAuthCheck, middleware.FileUploadRateLimit(33, 24*time.Hour), api.SaveImgByGinContext)
	// 文件获取接口 - 通过路径
	r.GET("/img/*filename", middleware.BrowserCache, api.GetFileByFileName)
}
