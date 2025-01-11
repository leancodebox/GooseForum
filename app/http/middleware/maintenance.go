package middleware

import (
	"net/http"

	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"

	"github.com/gin-gonic/gin"
)

var maintenanceHTML []byte = []byte(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8">
	<title>维护中</title>
	<style>
		body {
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh;
			margin: 0;
			background-color: #f2f2f2;
		}
		h1 {
			color: #333;
			font-family: Arial, sans-serif;
		}
	</style>
</head>
<body>
	<h1>当前网站正在维护中，请稍后再试。</h1>
</body>
</html>
`)

// SiteMaintenance 中间件用于检查站点是否处于维护状态
func SiteMaintenance(c *gin.Context) {
	// 从配置文件中读取维护模式状态
	maintenance := preferences.GetBool("app.enabled")
	if maintenance {
		// 设置HTTP状态码为503 Service Unavailable
		c.Writer.WriteHeader(http.StatusServiceUnavailable)

		// 定义维护页面的 HTML 内容

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
