package middleware

import (
	"github.com/leancodebox/goose/jwt"
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
		c.JSON(http.StatusUnauthorized, "未登陆")
		c.Abort()
		return
	}
	userId, newToken, err := jwt.VerifyTokenWithFresh(token)
	if err != nil {
		errorMsg := err.Error()
		if err == jwt.TokenExpired {
			errorMsg = "授权已过期"
		}
		c.JSON(http.StatusUnauthorized, errorMsg)
		c.Abort()
		return
	}
	if token != newToken {
		c.Header("New-Token", newToken)
	}
	c.Set("userId", userId)
	c.Next()
}
