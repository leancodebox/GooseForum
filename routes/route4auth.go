package routes

import (
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"

	"github.com/gin-gonic/gin"
)

func ginAuth(ginApp *gin.Engine) {
	ginApp.Group("api").
		POST("reg", ginUpP(controllers.Register)).
		POST("login", ginUpP(controllers.Login)).
		POST("get-captcha", ginUpNP(controllers.GetCaptcha))

	ginApp.Group("api").Use(middleware.JWTAuth4Gin).
		GET("get-user-info-v4", UpButterReq(controllers.UserInfo)).
		POST("set-user-info", UpButterReq(controllers.EditUserInfo)).
		POST("invitation", UpButterReq(controllers.Invitation))
}
