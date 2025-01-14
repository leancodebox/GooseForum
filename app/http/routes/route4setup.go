package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
)

// RegisterSetupRoutes 只注册setup相关的路由
func RegisterSetupRoutes(ginApp *gin.Engine) {
	setupGroup := ginApp.Group("api/setup")
	setupGroup.GET("status", UpButterReq(controllers.GetSetupStatus))
	setupGroup.POST("init", UpButterReq(controllers.InitialSetup))

	// 添加静态文件服务用于setup页面
	ginApp.Static("/", "./actor/dist")
}
