package routes

import (
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/bundles/app"

	"github.com/gin-contrib/gzip"

	"github.com/gin-gonic/gin"
)

func frontend(ginApp *gin.Engine) {
	actGroup := ginApp.Group("/actor")
	if app.IsProduction() {
		actGroup.Use(middleware.CacheMiddleware).
			Use(gzip.Gzip(gzip.DefaultCompression)).
			StaticFS("", PFilSystem("./actor/dist", app.GetActorFS()))
	} else {
		actGroup.Use(gzip.Gzip(gzip.DefaultCompression)).Static("", "./actor/dist")
	}
}
