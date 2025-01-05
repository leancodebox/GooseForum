package routes

import (
	"github.com/leancodebox/GooseForum/app/assert"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"path"
)

func auth(ginApp *gin.Engine) {
	ginApp.Group("api").
		POST("reg", ginUpP(controllers.Register)).
		POST("login", ginUpP(controllers.Login)).
		GET("get-captcha", ginUpNP(controllers.GetCaptcha)).
		POST("get-user-info-show", ginUpP(controllers.GetUserInfo))

	ginApp.Group("api").Use(middleware.JWTAuth4Gin).
		GET("get-user-info", UpButterReq(controllers.UserInfo)).
		POST("set-user-info", UpButterReq(controllers.EditUserInfo)).
		POST("invitation", UpButterReq(controllers.Invitation)).
		POST("upload-avatar", controllers.UploadAvatar)

	// 添加静态文件服务，用于访问头像
	avatarPath := path.Join(setting.GetStorage(), "avatars")
	ginApp.Static("api/avatars", avatarPath)
	ginApp.GET("/api/assets/default-avatar.png", func(context *gin.Context) {
		context.Data(http.StatusOK, "image/jpeg", assert.GetDefaultAvatar())
	})
}
