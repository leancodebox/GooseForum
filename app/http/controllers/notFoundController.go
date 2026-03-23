package controllers

import (
	"net/http"
	"path"
	"strings"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/leancodebox/GooseForum/resource"

	"github.com/gin-gonic/gin"
)

const (
	contentTypeHTML      = "text/html"
	errorCodeNotFound    = 404
	errorMessageNotFound = "路由未定义，请确认 url 和请求方法是否正确。"
)

func NotFound(c *gin.Context) {
	uri := c.Request.RequestURI
	// SPA 回退：凡是以 /actor 开头的未知路由，返回 index.html 交由前端路由处理
	if strings.HasPrefix(uri, "/admin") {
		fsEntity, _ := resource.GetAdminFS()
		// 不要使用index.html
		// Go net/http 有把 index.html 处理为 ./ 的奇怪逻辑
		c.FileFromFS(path.Join(""), http.FS(fsEntity))
		return
	}
	if strings.Contains(c.GetHeader("Accept"), contentTypeHTML) {
		c.Redirect(http.StatusTemporaryRedirect, urlconfig.Home())
		return
	}
	c.JSON(http.StatusNotFound, component.DataMap{
		"code": errorCodeNotFound,
		"msg":  errorMessageNotFound,
	})
}
