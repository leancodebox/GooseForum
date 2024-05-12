package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Api(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{
		"msg": "OK",
	})
}

const (
	contentTypeHTML      = "text/html"
	errorCodeNotFound    = 404
	errorMessageNotFound = "路由未定义，请确认 url 和请求方法是否正确。"
)

func NotFound(c *gin.Context) {
	acceptString := c.GetHeader("Accept")
	if strings.Contains(acceptString, contentTypeHTML) {
		c.Redirect(http.StatusTemporaryRedirect, "/actor")
		return
	}
	c.JSON(http.StatusNotFound, component.DataMap{
		"error_code":    errorCodeNotFound,
		"error_message": errorMessageNotFound,
	})
}

func About() component.Response {
	return component.SuccessResponse(component.DataMap{
		"message": "Hello~ Now you see a json from gin",
	})
}
