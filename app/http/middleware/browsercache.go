package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/httputil"
)

// BrowserCache applies long public cache headers in production.
func BrowserCache(c *gin.Context) {
	if !setting.IsProduction() {
		c.Next()
		return
	}
	httputil.SetLongPublic(c)
	c.Next()
}
