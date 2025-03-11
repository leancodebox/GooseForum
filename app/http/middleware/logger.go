package middleware

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLog(c *gin.Context) {
	if !preferences.GetBool("server.accessLog", false) {
		c.Next()
		return
	}
	startTime := time.Now()
	c.Next()
	endTime := time.Now()
	latencyTime := fmt.Sprintf("%6v", endTime.Sub(startTime))
	reqMethod := c.Request.Method
	reqUri := c.Request.RequestURI
	statusCode := c.Writer.Status()
	clientIP := c.ClientIP()
	slog.Info(
		"access",
		"http_status", statusCode,
		"total_time", latencyTime,
		"ip", clientIP,
		"method", reqMethod,
		"uri", reqUri,
	)
}
