package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/httputil"
)

func BrowserCache(c *gin.Context) {
	if !setting.IsProduction() {
		c.Next()
		return
	}
	if httputil.IsAdminIndexPath(c.Request.URL.Path) {
		httputil.SetNoStore(c)
		c.Next()
		return
	}
	httputil.SetLongPublic(c)
	c.Next()
}
