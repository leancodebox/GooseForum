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
			background: linear-gradient(135deg, #ff6f61, #6ab04c);
		}
		h1 {
			color: #333333;
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.3);
		}
		p {
			color: #333333;
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.2);
			margin-top: 10px;
		}
		.container {
			text-align: center;
			background-color: rgba(255, 255, 255, 0.8);
			padding: 40px;
			border-radius: 10px;
			box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>网站维护中</h1>
		<p>我们正在进行系统升级，请稍后再试。感谢您的耐心等待与支持！</p>
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
