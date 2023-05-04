package routes

import (
	"github.com/leancodebox/GooseForum/bundles/app"

	"github.com/gin-contrib/gzip"

	"github.com/gin-gonic/gin"
)

func ginWeb(ginApp *gin.Engine) {
	if app.IsProduction() {
		ginApp.Use(gzip.Gzip(gzip.DefaultCompression)).StaticFS("/actor", PFilSystem("./actor/dist", app.GetActorFS()))
	} else {
		ginApp.Use(gzip.Gzip(gzip.DefaultCompression)).Static("/actor", "./actor/dist")
	}
}
