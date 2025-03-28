package controllers

import (
	"github.com/leancodebox/GooseForum/app/assert"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contentTypeHTML      = "text/html"
	errorCodeNotFound    = 404
	errorMessageNotFound = "路由未定义，请确认 url 和请求方法是否正确。"
)

func NotFound(c *gin.Context) {
	uri := c.Request.RequestURI
	if !strings.HasPrefix(uri, "/app/assert") && strings.HasPrefix(uri, "/app") {
		fsEntity, _ := fs.Sub(assert.GetActorFs(), "frontend/dist2")
		if strings.HasPrefix(uri, "/app/admin") {
			c.FileFromFS(
				path.Join("admin.html"),
				http.FS(fsEntity),
			)
		} else {
			// 不要使用index.html
			// Go net/http 有把 index.html 处理为 ./ 的奇怪逻辑
			c.FileFromFS(
				path.Join(""),
				http.FS(fsEntity),
			)
		}
		return
	}
	acceptString := c.GetHeader("Accept")
	if strings.Contains(acceptString, contentTypeHTML) {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	c.JSON(http.StatusNotFound, component.DataMap{
		"code": errorCodeNotFound,
		"msg":  errorMessageNotFound,
	})
}
