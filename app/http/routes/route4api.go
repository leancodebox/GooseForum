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

func frontend(ginApp *gin.Engine) {
	actGroup := ginApp.Group("/")
	appFs, _ := fs.Sub(assert.GetActorFs(), "frontend/dist")
	staticFS, _ := resource.GetStaticFS()
	actGroup.Use(middleware.CacheMiddleware).
		Use(gzip.Gzip(gzip.DefaultCompression)).
		Use(middleware.BrowserCache).
		StaticFS("static", http.FS(staticFS)).
		StaticFS("app", http.FS(appFs))

	// SEO 相关路由
	ginApp.GET("/robots.txt", controllers.RenderRobotsTxt)
	ginApp.GET("/sitemap.xml", controllers.RenderSitemapXml)
	ginApp.GET("/rss.xml", controllers.RenderRssV2)
}

// 认证相关服务
func auth(ginApp *gin.Engine) {
	// 非登陆下的用户操作
	ginApp.Group("api").
		POST("reg", ginUpP(controllers.Register)).
		POST("login", ginUpP(controllers.Login)).
		GET("get-captcha", ginUpNP(controllers.GetCaptcha)).
		POST("get-user-info-show", ginUpP(controllers.GetUserInfo))
	// 登陆状态下的用户操作
	ginApp.Group("api").Use(middleware.JWTAuth4Gin).
		GET("get-user-info", UpButterReq(controllers.UserInfo)).
		POST("set-user-info", UpButterReq(controllers.EditUserInfo)).
		POST("invitation", UpButterReq(controllers.Invitation)).
		POST("upload-avatar", controllers.UploadAvatar).
		POST("change-password", UpButterReq(controllers.ChangePassword))
	// 添加激活路由
	ginApp.GET("api/activate", controllers.ActivateAccount)
}

func viewRoute(ginApp *gin.Engine) {
	view := ginApp.Group("")
	view.Use(middleware.JWTAuth)
	view.GET("", controllers.RenderArticlesPage)
	view.GET("/post", controllers.RenderArticlesPage)
	view.GET("/post/:id", controllers.RenderArticleDetail)
	view.GET("/login", controllers.LoginPage)
	view.POST("/login", controllers.LoginHandler)
	view.POST("/register", controllers.RegisterHandle)
	view.POST("/logout", controllers.Logout)
	view.GET("/user-profile/:id", controllers.UserProfile)
	view.GET("/sponsors", controllers.Sponsors)
}

func forum(ginApp *gin.Engine) {
	forumApi := ginApp.Group("api/bbs")
	// 站点统计
	forumApi.GET("get-site-statistics", ginUpNP(controllers.GetSiteStatistics))
	// 分类列表
	forumApi.GET("get-articles-enum", ginUpNP(controllers.GetArticlesEnum))
	forumApi.GET("get-articles-category", ginUpNP(controllers.GetArticlesCategory))

	loginApi := forumApi.Use(middleware.JWTAuth4Gin)
	// 通知相关接口
	loginApi.POST("notification/list", UpButterReq(controllers.GetNotificationList))
	loginApi.GET("notification/unread-count", UpButterReq(controllers.GetUnreadCount))
	loginApi.POST("notification/mark-read", UpButterReq(controllers.MarkAsRead))
	loginApi.POST("notification/mark-all-read", UpButterReq(controllers.MarkAllAsRead))
	loginApi.POST("notification/delete", UpButterReq(controllers.DeleteNotification))
	loginApi.GET("notification/types", UpButterReq(controllers.GetNotificationTypes))

	// 编辑文章时原始文章内容
	loginApi.POST("get-articles-origin", middleware.CheckLogin, UpButterReq(controllers.WriteArticlesOrigin))
	// 发布文章
	loginApi.POST("write-articles", UpButterReq(controllers.WriteArticles))
	// 回复文章
	loginApi.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	// 回复评论
	loginApi.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	// 用户文章列表
	loginApi.POST("get-user-articles", UpButterReq(controllers.GetUserArticles))
	// 文章点赞
	loginApi.POST("like-articles", UpButterReq(controllers.LikeArticle))

	adminApi := ginApp.Group("api/admin").Use(middleware.JWTAuth4Gin)
	adminApi.POST("user-list", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.UserList))
	adminApi.POST("user-edit", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.EditUser))
	adminApi.POST("get-all-role-item", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.GetAllRoleItem))
	adminApi.POST("articles-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.ArticlesList))
	adminApi.POST("article-edit", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.EditArticle))
	adminApi.POST("get-permission-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.GetPermissionList))
	adminApi.POST("role-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleList))
	adminApi.POST("role-save", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleSave))
	adminApi.POST("role-delete", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleDel))
	adminApi.POST("opt-record-page", middleware.CheckPermission(permission.Admin), UpButterReq(controllers.OptRecordPage))
	adminApi.POST("category-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.GetCategoryList))
	adminApi.POST("category-save", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.SaveCategory))
	adminApi.POST("category-delete", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.DeleteCategory))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	// 文件上传接口
	r.POST("/img-upload", middleware.JWTAuth4Gin, controllers.SaveFileByGinContext)
	// 文件获取接口 - 通过路径
	r.GET("/img/*filename", middleware.BrowserCache, controllers.GetFileByFileName)
}
