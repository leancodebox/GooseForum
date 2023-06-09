package routes

import (
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"

	"github.com/gin-gonic/gin"
)

func ginBBS(ginApp *gin.Engine) {
	bbs := ginApp.Group("api/bbs")
	// 文章列表
	bbs.POST("get-articles", ginUpP(controllers.GetArticles))
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
	// 发布评论
	bbsAuth.POST("articles-comment", ginUpP(controllers.ArticleComment))
	// 申请展示 todo
	bbsAuth.POST("apply-show", UpButterReq(controllers.ApplyShow))

}
