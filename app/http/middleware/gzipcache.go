package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"sync"
)

var cache sync.Map

type cachedResponse struct {
	contentType string
	body        []byte
}
type cachingResponseWriter struct {
	gin.ResponseWriter
	body *[]byte
}

func (w cachingResponseWriter) Write(b []byte) (int, error) {
	*w.body = append(*w.body, b...)
	return w.ResponseWriter.Write(b)
}

func CacheMiddleware(c *gin.Context) {
	// 如果浏览器支持 Gzip 那么就开启缓存，否则就直接执行下个中间件
	if acceptEncoding := c.Request.Header.Get("Accept-Encoding"); strings.Contains(acceptEncoding, "gzip") {
		key := c.Request.URL.Path + "?" + c.Request.URL.RawQuery
		// 检查缓存
		if val, ok := cache.Load(key); ok {
			cachedResp := val.(cachedResponse)
			c.Header("Content-Encoding", "gzip")
			c.Data(http.StatusOK, cachedResp.contentType, cachedResp.body)
			c.Abort()
			return
		}

		// 使用自定义的ResponseWriter
		var respBody []byte
		writer := cachingResponseWriter{c.Writer, &respBody}
		c.Writer = writer
		// 执行下一个handler
		c.Next()

		// 缓存响应
		if c.Writer.Status() == http.StatusOK {
			contentType := c.Writer.Header().Get("Content-Type")
			cache.Store(key, cachedResponse{contentType, respBody})
		}
	} else {
		c.Next()
	}

}
