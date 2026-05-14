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
	if strings.HasPrefix(uri, "/admin") {
		fsEntity, _ := resource.GetAdminFS()
		// Empty path serves the admin index without net/http rewriting index.html to ./.
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
