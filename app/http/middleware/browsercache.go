package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func BrowserCache(c *gin.Context) {
	// 如果请求的是 index.html，禁用浏览器缓存
	if c.Request.URL.Path == "/actor/" {
		c.Header("Cache-Control", "no-store")
	} else if strings.HasPrefix(c.Request.RequestURI, "/app") {
		c.Header("Cache-Control", "no-store")
	} else {
		// 其他情况下继续使用浏览器缓存
		c.Header("Cache-Control", "public, max-age=3600")
	}
	c.Next()
}
