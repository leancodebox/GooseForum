package middleware

import (
	"github.com/gin-gonic/gin"
)

func SiteInfo(context *gin.Context) {
	context.Header("X-Powered-By", "GooseForum/0.0.1")
	context.Next()
}
