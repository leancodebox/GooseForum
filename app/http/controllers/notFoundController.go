package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contentTypeHTML      = "text/html"
	errorCodeNotFound    = 404
	errorMessageNotFound = "路由未定义，请确认 url 和请求方法是否正确。"
)

func NotFound(c *gin.Context) {
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
