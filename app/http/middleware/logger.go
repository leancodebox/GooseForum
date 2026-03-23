package middleware

import (
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"

	"github.com/gin-gonic/gin"
)

func AccessLog(c *gin.Context) {
	if !preferences.GetBool("server.accessLog", false) {
		c.Next()
		return
	}

	startTime := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	c.Next()

	latency := time.Since(startTime)
	statusCode := c.Writer.Status()

	if raw != "" {
		path = path + "?" + raw
	}

	// 构造基础日志字段
	fields := []any{
		"status", statusCode,
		"latency", latency.String(),
		"latency_ms", latency.Milliseconds(), // 方便后续数字化分析
		"ip", c.ClientIP(),
		"method", c.Request.Method,
		"path", path,
		"route", c.FullPath(), // 记录参数化路由，如 /api/user/:id
	}

	// 如果有 Gin 内部错误，记录下来
	slog.Info("access", fields...)

}
