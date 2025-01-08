package middleware

import (
	jwt "github.com/leancodebox/GooseForum/app/bundles/goose/jwtopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth4Gin(c *gin.Context) {
	var token string
	//token = c.GetHeader("x-token")
	token = c.GetHeader("Authorization")
	token = strings.ReplaceAll(token, "Bearer ", "")
	if token == "" {
		c.JSON(http.StatusUnauthorized, component.FailData("未登陆"))
		c.Abort()
		return
	}
	userId, newToken, err := jwt.VerifyTokenWithFresh(token)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(http.StatusUnauthorized, component.FailData(errorMsg))
		c.Abort()
		return
	}
	if token != newToken {
		c.Header("New-Token", newToken)
	}
	c.Set("userId", userId)
	c.Next()
}
