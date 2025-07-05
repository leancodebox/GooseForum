package routes

import (
	"net/http"

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
	viewRouteApp.GET("/admin/*path", middleware.CheckPermissionOrNoUser(permission.Admin), controllers.Admin)
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
	baseApi.GET("get-captcha", ginUpNP(controllers.GetCaptcha))
	// 添加激活路由
	baseApi.GET("activate", controllers.ActivateAccount)

	// 登陆状态下的用户操作
	loginApi := ginApp.Group("api").Use(middleware.JWTAuth4Gin)
	// 用户信息
	loginApi.GET("get-user-info", UpButterReq(controllers.UserInfo))
	// 设置用户信息
	loginApi.POST("set-user-info", UpButterReq(controllers.EditUserInfo))
	// 邀请码
	loginApi.POST("invitation", UpButterReq(controllers.Invitation))
	// 上传头像
	loginApi.POST("upload-avatar", controllers.UploadAvatar)
	// 修改密码
	loginApi.POST("change-password", UpButterReq(controllers.ChangePassword))

	forumApi := baseApi.Group("forum")
	forumApi.POST("apply-link-add", UpButterReq(controllers.ApplyAddLink))
	// 站点统计
	forumApi.GET("get-site-statistics", ginUpNP(controllers.GetSiteStatistics))
	// 分类列表
	forumApi.GET("get-articles-enum", ginUpNP(controllers.GetArticlesEnum))

	forumLoginApi := forumApi.Use(middleware.JWTAuth4Gin)
	// 通知相关接口
	forumLoginApi.POST("notification/list", UpButterReq(controllers.GetNotificationList))
	forumLoginApi.POST("notification/query", UpButterReq(controllers.QueryNotificationList))
	forumLoginApi.GET("notification/unread-count", UpButterReq(controllers.GetUnreadCount))
	forumLoginApi.GET("notification/last-unread", UpButterReq(controllers.GetLastUnread))
	forumLoginApi.POST("notification/mark-read", UpButterReq(controllers.MarkAsRead))
	forumLoginApi.POST("notification/mark-all-read", UpButterReq(controllers.MarkAllAsRead))
	forumLoginApi.POST("notification/delete", UpButterReq(controllers.DeleteNotification))
	forumLoginApi.GET("notification/types", UpButterReq(controllers.GetNotificationTypes))
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

	adminApi := baseApi.Group("admin", middleware.JWTAuth4Gin)

	// 用户管理
	adminApi.
		Group("", middleware.CheckPermission(permission.UserManager)).
		POST("user-list", UpButterReq(controllers.UserList)).
		POST("user-edit", UpButterReq(controllers.EditUser)).
		POST("get-all-role-item", UpButterReq(controllers.GetAllRoleItem))

	// 文章管理 &  分类管理
	adminApi.Group("", middleware.CheckPermission(permission.ArticlesManager)).
		POST("articles-list", UpButterReq(controllers.ArticlesList)).
		POST("article-edit", UpButterReq(controllers.EditArticle)).
		POST("category-list", UpButterReq(controllers.GetCategoryList)).
		POST("category-save", UpButterReq(controllers.SaveCategory)).
		POST("category-delete", UpButterReq(controllers.DeleteCategory))

	// 权限管理
	adminApi.Group("", middleware.CheckPermission(permission.RoleManager)).
		POST("get-permission-list", UpButterReq(controllers.GetPermissionList)).
		POST("role-list", UpButterReq(controllers.RoleList)).
		POST("role-save", UpButterReq(controllers.RoleSave)).
		POST("role-delete", UpButterReq(controllers.RoleDel))

	// 操作日志
	adminApi.Group("", middleware.CheckPermission(permission.Admin)).
		POST("opt-record-page", UpButterReq(controllers.OptRecordPage))

	// 站点管理
	adminApi.Group("", middleware.CheckPermission(permission.SiteManager)).
		POST("apply-sheet-list", UpButterReq(controllers.ApplySheet)).
		GET("friend-links", UpButterReq(controllers.GetFriendLinks)).
		POST("save-friend-links", UpButterReq(controllers.SaveFriendLinks))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	// 文件上传接口
	r.POST("/img-upload", middleware.JWTAuth4Gin, controllers.SaveFileByGinContext)
	// 文件获取接口 - 通过路径
	r.GET("/img/*filename", middleware.BrowserCache, controllers.GetFileByFileName)
}
