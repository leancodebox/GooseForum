package routes

import (
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/api"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"net/http"
	"time"

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
	ginApp.Group("/").
		Use(middleware.CacheMiddleware).
		Use(gzip.Gzip(gzip.DefaultCompression)).
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
		viewrender.Reload()
		c.String(200, "模板已刷新")
	})
	ginApp.POST("/login", controllers.Login)
	ginApp.POST("/register", controllers.Register)
	ginApp.POST("/logout", controllers.Logout)

	viewRouteApp := ginApp.Group("")
	viewRouteApp.Use(middleware.JWTAuth).
		Use(gzip.Gzip(gzip.DefaultCompression))
	viewRouteApp.GET("", controllers.Home)
	viewRouteApp.GET("/login", middleware.CheckNeedLogin, controllers.LoginView)
	viewRouteApp.GET("/user/:id", controllers.User)
	viewRouteApp.GET("/post", controllers.Post)
	viewRouteApp.GET("/post/:id", controllers.PostDetail)
	viewRouteApp.GET("/about", controllers.About)
	viewRouteApp.GET("/sponsors", controllers.SponsorsView)
	viewRouteApp.GET("/links", controllers.LinksView)
	viewRouteApp.GET("/terms-of-service", controllers.TermsOfService)
	viewRouteApp.GET("/privacy-policy", controllers.PrivacyPolicy)
	viewRouteApp.GET("/profile", middleware.CheckLogin, controllers.Profile)
	viewRouteApp.GET("/publish", middleware.CheckLogin, controllers.PublishV3)
	viewRouteApp.GET("/notifications", middleware.CheckLogin, controllers.Notifications)
	viewRouteApp.GET("/search", controllers.SearchPage)
	viewRouteApp.GET("/admin/*path", middleware.CheckPermissionOrNoUser(permission.Admin), controllers.Admin)

	// 文档相关路由
	viewRouteApp.GET("/docs", controllers.DocsHome)
	viewRouteApp.GET("/docs/:project", controllers.DocsVersion)
	viewRouteApp.GET("/docs/:project/:version", controllers.DocsVersion)
	viewRouteApp.GET("/docs/:project/:version/:content", controllers.DocsContent)
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
	// 验证码
	baseApi.GET("get-captcha", ginUpNP(api.GetCaptcha))
	// 添加激活路由
	baseApi.GET("activate", controllers.ActivateAccount)

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
	// 回复文章
	forumLoginApi.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	// 回复评论
	forumLoginApi.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	// 用户文章列表
	forumLoginApi.POST("get-user-articles", UpButterReq(controllers.GetUserArticles))
	// 文章点赞
	forumLoginApi.POST("like-articles", UpButterReq(controllers.LikeArticle))
	// 关注
	forumLoginApi.POST("follow-user", UpButterReq(controllers.FollowUser))

	adminApi := baseApi.Group("admin", middleware.JWTAuthCheck)

	// 用户管理
	adminApi.
		Group("", middleware.CheckPermission(permission.UserManager)).
		POST("user-list", UpButterReq(api.UserList)).
		POST("user-edit", UpButterReq(api.EditUser)).
		POST("get-all-role-item", UpButterReq(api.GetAllRoleItem))

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
		GET("friend-links", UpButterReq(api.GetFriendLinks)).
		POST("save-friend-links", UpButterReq(api.SaveFriendLinks)).
		GET("web-settings", UpButterReq(api.GetWebSettings)).
		POST("save-web-settings", UpButterReq(api.SaveWebSettings)).
		GET("site-settings", UpButterReq(api.GetSiteSettings)).
		POST("save-site-settings", UpButterReq(api.SaveSiteSettings)).
		GET("mail-settings", UpButterReq(api.GetMailSettings)).
		POST("save-mail-settings", UpButterReq(api.SaveMailSettings)).
		POST("test-mail-connection", UpButterReq(api.TestMailConnection)).
		GET("footer-links", UpButterReq(api.GetFooterLinks)).
		POST("save-footer-links", UpButterReq(api.SaveFooterLinks)).
		GET("sponsors", UpButterReq(api.GetSponsors)).
		POST("save-sponsors", UpButterReq(api.SaveSponsors))

	// 文档管理
	adminApi.Group("", middleware.CheckPermission(permission.Admin)).
		POST("docs/projects/list", UpButterReq(api.AdminDocsProjectList)).
		GET("docs/projects/:id", UpButterReq(api.AdminDocsProjectDetail)).
		POST("docs/projects", UpButterReq(api.AdminDocsProjectCreate)).
		PUT("docs/projects/:id", UpButterReq(api.AdminDocsProjectUpdate)).
		DELETE("docs/projects/:id", UpButterReq(api.AdminDocsProjectDelete)).
		POST("docs/versions/list", UpButterReq(api.AdminDocsVersionList)).
		GET("docs/versions/:id", UpButterReq(api.AdminDocsVersionDetail)).
		POST("docs/versions", UpButterReq(api.AdminDocsVersionCreate)).
		PUT("docs/versions/:id", UpButterReq(api.AdminDocsVersionUpdate)).
		DELETE("docs/versions/:id", UpButterReq(api.AdminDocsVersionDelete)).
		PUT("docs/versions/:id/set-default", UpButterReq(api.AdminDocsVersionSetDefault)).
		PUT("docs/versions/:id/directory", UpButterReq(api.AdminDocsVersionDirectoryUpdate)).
		POST("docs/contents/list", UpButterReq(api.AdminDocsContentList)).
		GET("docs/contents/:id", UpButterReq(api.AdminDocsContentDetail)).
		POST("docs/contents", UpButterReq(api.AdminDocsContentCreate)).
		PUT("docs/contents/:id", UpButterReq(api.AdminDocsContentUpdate)).
		DELETE("docs/contents/:id", UpButterReq(api.AdminDocsContentDelete)).
		POST("docs/contents/:id/publish", UpButterReq(api.AdminDocsContentPublish)).
		POST("docs/contents/:id/draft", UpButterReq(api.AdminDocsContentDraft)).
		POST("docs/contents/preview", UpButterReq(api.AdminDocsContentPreview))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	// 文件上传接口 - 每日最多上传n张图片
	r.POST("/img-upload", middleware.JWTAuthCheck, middleware.FileUploadRateLimit(33, 24*time.Hour), api.SaveImgByGinContext)
	// 文件获取接口 - 通过路径
	r.GET("/img/*filename", middleware.BrowserCache, api.GetFileByFileName)
}
