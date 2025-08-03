package middleware

import (
	_ "embed"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed maintenance.html
var maintenanceHTML []byte

// SiteMaintenance 中间件用于检查站点是否处于维护状态
func SiteMaintenance(c *gin.Context) {
	// 从配置文件中读取维护模式状态
	maintenance := preferences.GetBool("app.maintenance")
	if maintenance {
		// 设置HTTP状态码为503 Service Unavailable
		c.Writer.WriteHeader(http.StatusServiceUnavailable)

		// 返回优化后的维护页面
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Writer.Write(maintenanceHTML)
		// 中止后续处理
		c.Abort()
		return
	}
	// 如果不在维护模式，则继续处理请求
	c.Next()

}
