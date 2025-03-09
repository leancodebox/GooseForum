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

func setup(ginApp *gin.Engine) {
	setupGroup := ginApp.Group("api/setup")
	setupGroup.GET("status", UpButterReq(controllers.GetSetupStatus))
	setupGroup.POST("init", UpButterReq(controllers.InitialSetup))

}

func frontend(ginApp *gin.Engine) {
	actGroup := ginApp.Group("/")
	actorFs, _ := fs.Sub(assert.GetActorFs(), "frontend/dist")
	staticFS, _ := resource.GetStaticFS()
	actGroup.Use(middleware.CacheMiddleware).
		Use(gzip.Gzip(gzip.DefaultCompression)).
		StaticFS("actor", http.FS(actorFs)).
		StaticFS("static", http.FS(staticFS))

	// SEO 相关路由
	ginApp.GET("/robots.txt", controllers.RenderRobotsTxt)
	ginApp.GET("/sitemap.xml", controllers.RenderSitemapXml)
	ginApp.GET("/rss.xml", controllers.RenderRssFeed)
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

	// 静态头像地址
	ginApp.GET("/api/assets/default-avatar.png", func(context *gin.Context) {
		context.Data(http.StatusOK, "image/png", assert.GetDefaultAvatar())
	})
	// 添加激活路由
	ginApp.GET("api/activate", controllers.ActivateAccount)
}

func viewRoute(ginApp *gin.Engine) {
	view := ginApp.Group("")
	view.Use(middleware.JWTAuth)
	view.GET("", controllers.RenderIndex)
	view.GET("/post", controllers.RenderArticlesPage)
	view.GET("/post/:id", controllers.RenderArticleDetail)
	view.GET("/login", controllers.LoginPage)
	view.POST("/login", controllers.LoginHandler)
	view.GET("/notifications", middleware.CheckLogin, controllers.Notifications)
	view.GET("/post-edit", middleware.CheckLogin, controllers.PostEdit)
	view.GET("/user-profile", controllers.UserProfile)
	view.GET("/sponsors", controllers.Sponsors)
}

func bbs(ginApp *gin.Engine) {
	bbsShow := ginApp.Group("api/bbs")

	// 站点统计
	bbsShow.GET("get-site-statistics", ginUpNP(controllers.GetSiteStatistics))
	// 分类列表
	bbsShow.GET("get-articles-enum", ginUpNP(controllers.GetArticlesEnum))
	bbsShow.GET("get-articles-category", ginUpNP(controllers.GetArticlesCategory))
	// 文章分页
	bbsShow.POST("get-articles-page", ginUpP(controllers.GetArticlesPage))
	// 文章详情
	bbsShow.POST("get-articles-detail", ginUpP(controllers.GetArticlesDetail))

	// 热门链接
	// 用户主页
	// tag/分类

	bbsAuth := bbsShow.Use(middleware.JWTAuth4Gin)
	// 通知相关接口
	bbsAuth.POST("notification/list", UpButterReq(controllers.GetNotificationList))
	bbsAuth.GET("notification/unread-count", UpButterReq(controllers.GetUnreadCount))
	bbsAuth.POST("notification/mark-read", UpButterReq(controllers.MarkAsRead))
	bbsAuth.POST("notification/mark-all-read", UpButterReq(controllers.MarkAllAsRead))
	bbsAuth.POST("notification/delete", UpButterReq(controllers.DeleteNotification))
	bbsAuth.GET("notification/types", UpButterReq(controllers.GetNotificationTypes))

	// 编辑文章时原始文章内容
	bbsAuth.POST("get-articles-origin", UpButterReq(controllers.WriteArticlesOrigin))
	// 发布文章
	bbsAuth.POST("write-articles", UpButterReq(controllers.WriteArticles))
	// 回复文章
	bbsAuth.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	// 回复评论
	bbsAuth.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	// 申请展示 todo
	bbsAuth.POST("apply-show", UpButterReq(controllers.ApplyShow))
	// 用户文章列表
	bbsAuth.POST("/get-user-articles", UpButterReq(controllers.GetUserArticles))

	admin := ginApp.Group("api/admin").Use(middleware.JWTAuth4Gin)
	admin.POST("user-list", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.UserList))
	admin.POST("user-edit", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.EditUser))
	admin.POST("get-all-role-item", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.GetAllRoleItem))
	admin.POST("articles-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.ArticlesList))
	admin.POST("article-edit", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.EditArticle))
	admin.POST("get-permission-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.GetPermissionList))
	admin.POST("role-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleList))
	admin.POST("role-save", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleSave))
	admin.POST("role-delete", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleDel))
	admin.POST("opt-record-page", middleware.CheckPermission(permission.Admin), UpButterReq(controllers.OptRecordPage))
	admin.POST("category-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.GetCategoryList))
	admin.POST("category-save", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.SaveCategory))
	admin.POST("category-delete", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.DeleteCategory))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	// 文件上传接口
	r.POST("/img-upload", middleware.JWTAuth4Gin, controllers.SaveFileByGinContext)
	// 文件获取接口 - 通过路径
	r.GET("/img/*filename", controllers.GetFileByFileName)
}
