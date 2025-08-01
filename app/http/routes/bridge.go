package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
)

func RegisterByGin(ginApp *gin.Engine) {
	// 基础中间件
	ginApp.Use(middleware.SiteMaintenance)
	ginApp.Use(middleware.SiteInfo)
	ginApp.Use(middleware.GinCors)

	// 访问日志中间件
	ginApp.Use(middleware.AccessLog)

	siteInfoRoute(ginApp)
	// 接口
	apiRoute(ginApp)
	// 文件
	fileServer(ginApp)
	// view
	viewRoute(ginApp)
	// 资源
	assertRouter(ginApp)

	ginApp.NoRoute(controllers.NotFound)

}
