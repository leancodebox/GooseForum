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
	ginApp.GET("/user/:id", controllers.User)
	ginApp.GET("/post-v2", controllers.PostV2)
	ginApp.GET("/post-v2/:id", controllers.PostDetail)
}
