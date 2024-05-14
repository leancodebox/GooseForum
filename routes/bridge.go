package routes

import (
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterByGin(ginApp *gin.Engine) {
	// 基础中间件
	ginApp.Use(middleware.TraceInit)
	ginApp.Use(middleware.GinCors)

	// 前端资源
	ginWeb(ginApp)

	// 访问日志中间件
	ginApp.Use(middleware.GinLogger)

	// 接口
	ginApi(ginApp)
	ginAuth(ginApp)
	ginBBS(ginApp)

	ginApp.NoRoute(controllers.NotFound)
}
