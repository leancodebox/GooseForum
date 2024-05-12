package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
)

func ginApi(ginApp *gin.Engine) {
	ginApp.GET("/api", controllers.Api)

	apiGroup := ginApp.Group("api")
	apiGroup.GET("about", ginUpNP(controllers.About))

}
