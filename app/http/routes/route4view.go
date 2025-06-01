package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
)

func Reload() {
	viewrender.Reload()
}

func view(ginApp *gin.Engine) {
	ginApp.GET("/post-v2", controllers.PostV2)
}
