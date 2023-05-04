package routes

import (
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterByGin(ginApp *gin.Engine) {

	ginApp.Use(middleware.TraceInit)
	ginApp.Use(middleware.GinCors)
	ginApp.Use(middleware.GinLogger)

	ginWeb(ginApp)
	ginApi(ginApp)
	ginAuth(ginApp)
	ginBBS(ginApp)

	ginApp.NoRoute(controllers.NotFound)
}
