package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/assert"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"net/http"
	"path"
)

func setup(ginApp *gin.Engine) {
	setupGroup := ginApp.Group("api/setup")
	setupGroup.GET("status", UpButterReq(controllers.GetSetupStatus))
	setupGroup.POST("init", UpButterReq(controllers.InitialSetup))
}

func auth(ginApp *gin.Engine) {
	ginApp.Group("api").
		POST("reg", ginUpP(controllers.Register)).
		POST("login", ginUpP(controllers.Login)).
		GET("get-captcha", ginUpNP(controllers.GetCaptcha)).
		POST("get-user-info-show", ginUpP(controllers.GetUserInfo))

	ginApp.Group("api").Use(middleware.JWTAuth4Gin).
		GET("get-user-info", UpButterReq(controllers.UserInfo)).
		POST("set-user-info", UpButterReq(controllers.EditUserInfo)).
		POST("invitation", UpButterReq(controllers.Invitation)).
		POST("upload-avatar", controllers.UploadAvatar)

	// 添加静态文件服务，用于访问头像
	avatarPath := path.Join(setting.GetStorage(), "avatars")
	ginApp.Static("api/avatars", avatarPath)
	ginApp.GET("/api/assets/default-avatar.png", func(context *gin.Context) {
		context.Data(http.StatusOK, "image/png", assert.GetDefaultAvatar())
	})
}

func bbs(ginApp *gin.Engine) {
	bbsShow := ginApp.Group("api/bbs")

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

	bbsAuth := bbsShow.Use(middleware.JWTAuth4Gin).
		Use(middleware.JWTAuth4Gin)
	// 通知相关接口
	bbsAuth.POST("notification/list", UpButterReq(controllers.GetNotificationList))
	bbsAuth.GET("notification/unread-count", UpButterReq(controllers.GetUnreadCount))
	bbsAuth.POST("notification/mark-read", UpButterReq(controllers.MarkAsRead))
	bbsAuth.POST("notification/mark-all-read", UpButterReq(controllers.MarkAllAsRead))
	bbsAuth.POST("notification/delete", UpButterReq(controllers.DeleteNotification))
	bbsAuth.GET("notification/types", UpButterReq(controllers.GetNotificationTypes))

	// 其他接口...
	bbsAuth.POST("get-articles-origin", UpButterReq(controllers.WriteArticlesOrigin))
	// 发布文章
	bbsAuth.POST("write-articles", UpButterReq(controllers.WriteArticles))
	// 回复文章
	bbsAuth.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	// 回复评论
	bbsAuth.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	// 申请展示 todo
	bbsAuth.POST("apply-show", UpButterReq(controllers.ApplyShow))
	//
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
