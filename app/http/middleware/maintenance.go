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
			background: linear-gradient(135deg, #71b7e6, #9b59b6);
		}
		h1 {
			color: #fff;
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
		}
		p {
			color: #f0f0f0;
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.3);
			margin-top: 20px;
		}
		.container {
			text-align: center;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>抱歉，我们正在进行维护</h1>
		<p>请稍后再试。感谢您的理解与支持！</p>
	</div>
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
