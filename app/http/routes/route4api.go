package routes

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/assert"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/resource"
	"io/fs"
	"net/http"
)

func assertRouter(ginApp *gin.Engine) {
	appFs, _ := fs.Sub(assert.GetActorFs(), "frontend/dist")
	assetsFs, _ := fs.Sub(resource.GetViewAssert(), "static/dist/assets")
	staticFS, _ := resource.GetStaticFS()
	ginApp.Group("/").
		Use(middleware.CacheMiddleware).
		Use(gzip.Gzip(gzip.DefaultCompression)).
		Use(middleware.BrowserCache).
		StaticFS("assets", http.FS(assetsFs)).
		StaticFS("static", http.FS(staticFS)).
		StaticFS("app", http.FS(appFs))
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
	viewRouteApp.GET("/profile", middleware.CheckLogin, controllers.Profile)
	viewRouteApp.GET("/publish", middleware.CheckLogin, controllers.Publish)
	viewRouteApp.GET("/notifications", middleware.CheckLogin, controllers.Notifications)
	viewRouteApp.GET("/submit-link", controllers.SubmitLink)
}

func siteInfoRoute(ginApp *gin.Engine) {
	// SEO 相关路由
	ginApp.GET("/robots.txt", controllers.RenderRobotsTxt)
	ginApp.GET("/sitemap.xml", controllers.RenderSitemapXml)
	ginApp.GET("/rss.xml", controllers.RenderRssV2)
}

// 认证相关服务
func authApi(ginApp *gin.Engine) {
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

	adminApi := baseApi.Group("admin").Use(middleware.JWTAuth4Gin)

	// 用户管理
	adminApi.POST("user-list", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.UserList))
	adminApi.POST("user-edit", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.EditUser))
	adminApi.POST("get-all-role-item", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.GetAllRoleItem))

	// 文章管理
	adminApi.POST("articles-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.ArticlesList))
	adminApi.POST("article-edit", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.EditArticle))

	// 权限管理
	adminApi.POST("get-permission-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.GetPermissionList))
	adminApi.POST("role-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleList))
	adminApi.POST("role-save", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleSave))
	adminApi.POST("role-delete", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleDel))

	// 操作日志
	adminApi.POST("opt-record-page", middleware.CheckPermission(permission.Admin), UpButterReq(controllers.OptRecordPage))

	// 分类管理
	adminApi.POST("category-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.GetCategoryList))
	adminApi.POST("category-save", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.SaveCategory))
	adminApi.POST("category-delete", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.DeleteCategory))

	// 站点管理
	adminApi.POST("apply-sheet-list", middleware.CheckPermission(permission.SiteManager), UpButterReq(controllers.ApplySheet))
	adminApi.GET("friend-links", middleware.CheckPermission(permission.SiteManager), UpButterReq(controllers.GetFriendLinks))
	adminApi.POST("save-friend-links", middleware.CheckPermission(permission.SiteManager), UpButterReq(controllers.SaveFriendLinks))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	// 文件上传接口
	r.POST("/img-upload", middleware.JWTAuth4Gin, controllers.SaveFileByGinContext)
	// 文件获取接口 - 通过路径
	r.GET("/img/*filename", middleware.BrowserCache, controllers.GetFileByFileName)
}
