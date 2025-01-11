package routes

import (
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterByGin(ginApp *gin.Engine) {
	// 基础中间件
	ginApp.Use(middleware.SiteMaintenance)
	ginApp.Use(middleware.SiteInfo)
	ginApp.Use(middleware.GinCors)

	// 前端资源
	frontend(ginApp)

	// 访问日志中间件
	ginApp.Use(middleware.GinLogger)

	// 接口
	setup(ginApp)
	api(ginApp)
	auth(ginApp)
	bbs(ginApp)
	fileServer(ginApp)

	ginApp.NoRoute(controllers.NotFound)
}
