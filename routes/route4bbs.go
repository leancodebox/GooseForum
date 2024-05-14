package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

func ginBBS(ginApp *gin.Engine) {
	bbs := ginApp.Group("api/bbs")
	// 文章分页
	bbs.POST("get-articles-page", ginUpP(controllers.GetArticlesPage))
	// 文章详情
	bbs.POST("get-articles-detail", ginUpP(controllers.GetArticlesDetail))
	// 热门链接
	// 用户主页
	// tag/分类

	bbsAuth := bbs.Use(middleware.JWTAuth4Gin)
	// 发布文章
	bbsAuth.POST("write-articles", UpButterReq(controllers.WriteArticles))
	// 回复文章
	bbsAuth.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	// 回复评论
	bbsAuth.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	// 申请展示 todo
	bbsAuth.POST("apply-show", UpButterReq(controllers.ApplyShow))

	admin := ginApp.Group("api/admin").Use(middleware.JWTAuth4Gin)
	admin.POST("user-list", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.UserList))
	admin.POST("user-edit", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.EditUser))
	admin.POST("articles-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.ArticlesList))
	admin.POST("article-edit", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.EditArticle))
	admin.POST("role-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.RoleList))
	admin.POST("role-save", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.RoleSave))

}
