package routes

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/resourcev2"
	"io/fs"
	"net/http"
)

func Reload() {
	viewrender.Reload()
}

func viewAssert(ginApp *gin.Engine) {
	actGroup := ginApp.Group("/")
	appFs, _ := fs.Sub(resourcev2.GetViewAssert(), "static/dist/assets")
	actGroup.Use(middleware.CacheMiddleware).
		Use(gzip.Gzip(gzip.DefaultCompression)).
		Use(middleware.BrowserCache).
		StaticFS("assets", http.FS(appFs))
}

func view(ginApp *gin.Engine) {
	viewRouteApp := ginApp.Group("")
	viewRouteApp.Use(middleware.JWTAuth).
		Use(gzip.Gzip(gzip.DefaultCompression))
	viewRouteApp.GET("", controllers.Home)
	viewRouteApp.GET("/login", middleware.CheckNeedLogin, controllers.LoginView)
	viewRouteApp.GET("/user/:id", controllers.User)
	viewRouteApp.GET("/post", controllers.PostV2)
	viewRouteApp.GET("/post/:id", controllers.PostDetail)
	viewRouteApp.GET("/about", controllers.About)
	viewRouteApp.GET("/sponsors", controllers.SponsorsView)
	viewRouteApp.GET("/links", controllers.LinksView)
	viewRouteApp.GET("/profile", middleware.CheckLogin, controllers.Profile)
	viewRouteApp.GET("/publish", middleware.CheckLogin, controllers.Publish)
	viewRouteApp.GET("/notifications", middleware.CheckLogin, controllers.Notifications)
	viewRouteApp.GET("/submit-link", controllers.SubmitLink)
}
