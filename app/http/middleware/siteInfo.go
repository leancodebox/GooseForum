package middleware

import (
	"github.com/gin-gonic/gin"
)

// SiteInfo writes a simple X-Powered-By response header.
func SiteInfo(context *gin.Context) {
	context.Header("X-Powered-By", "GooseForum/0.0.1")
	context.Next()
}
