package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func GinLogger(c *gin.Context) {
	startTime := time.Now()

	c.Next()

	endTime := time.Now()
	latencyTime := fmt.Sprintf("%6v", endTime.Sub(startTime))
	reqMethod := c.Request.Method
	reqUri := c.Request.RequestURI
	statusCode := c.Writer.Status()
	clientIP := c.ClientIP()

	slog.Info("access",
		"httpStatus", statusCode,
		"latencyTime", latencyTime,
		"clientIP", clientIP,
		"method", reqMethod,
		"uri", reqUri,
	)

}
