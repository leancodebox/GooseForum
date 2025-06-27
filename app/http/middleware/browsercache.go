package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
)

func BrowserCache(c *gin.Context) {
	if !setting.IsProduction() {
		c.Next()
		return
	}
	c.Header("Cache-Control", "public, max-age=18144000")
	c.Next()
}
