package routes

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/assert"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/middleware"
)

func frontend(ginApp *gin.Engine) {
	actGroup := ginApp.Group("/actor")
	if setting.IsProduction() {
		actGroup.Use(middleware.CacheMiddleware).
			Use(gzip.Gzip(gzip.DefaultCompression)).
			StaticFS("", PFilSystem("./frontend/dist", assert.GetActorFs()))
	} else {
		actGroup.Use(gzip.Gzip(gzip.DefaultCompression)).Static("", "./actor/dist")
	}
}
