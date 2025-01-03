package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

func bbs(ginApp *gin.Engine) {
	bbsShow := ginApp.Group("api/bbs")

	// 分类列表
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
	bbsAuth.POST("get-notification", UpButterReq(controllers.GetNotification))
	//
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

}
