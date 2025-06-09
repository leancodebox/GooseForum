package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"net/http"
	"strings"
	"sync"
)

var gzipCache sync.Map

type cachedResponse struct {
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
}
type cachingResponseWriter struct {
	gin.ResponseWriter
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
}

func (w *cachingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.headers = w.Header().Clone() // Clone headers to avoid concurrent map writes
	w.ResponseWriter.WriteHeader(code)
}

func (w *cachingResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}
func CacheMiddleware(c *gin.Context) {
	if !setting.IsProduction() {
		c.Next()
		return
	}
	// 如果浏览器支持 Gzip 那么就开启缓存，否则就直接执行下个中间件
	if acceptEncoding := c.Request.Header.Get("Accept-Encoding"); strings.Contains(acceptEncoding, "gzip") {
		key := c.Request.URL.Path
		// 检查缓存
		if val, ok := gzipCache.Load(key); ok {
			if cachedResp, ok := val.(cachedResponse); ok {
				// 恢复响应头和状态码
				for k, vv := range cachedResp.headers {
					for _, v := range vv {
						c.Header(k, v)
					}
				}

				c.Data(cachedResp.statusCode, cachedResp.headers.Get("Content-Type"), cachedResp.body.Bytes())
				c.Abort()
				return
			}
		}

		// 使用自定义ResponseWriter
		writer := &cachingResponseWriter{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK, // 初始状态码，可能会被WriteHeader修改
			headers:        make(http.Header),
			body:           bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer
		c.Next()

		// 缓存响应
		if writer.statusCode == http.StatusOK {
			gzipCache.Store(key, cachedResponse{
				statusCode: writer.statusCode,
				headers:    writer.headers.Clone(), // Clone headers to avoid modification after storing
				body:       writer.body,
			})
		}
	} else {
		c.Next()
	}
}
