package middleware

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"net/http"

	"github.com/gin-gonic/gin"
	jwt "github.com/leancodebox/GooseForum/app/bundles/jwtopt"
)

func JWTAuthCheck(c *gin.Context) {
	JWTAuth(c)
	if c.GetUint64("userId") == 0 {
		c.JSON(http.StatusUnauthorized, component.FailData("not authorized"))
		c.Abort()
		return
	}
	c.Next()
}

func JWTAuth(c *gin.Context) {
	token := jwt.GetGinAccessToken(c)
	if token == "" {
		c.Next()
		return
	}
	userId, newToken, err := jwt.VerifyTokenWithFresh(token)
	if err != nil {
		c.Next()
		return
	}
	if token != newToken {
		jwt.TokenSetting(c, newToken)
	}
	c.Set("userId", userId)
	userservice.UpdateUserActivity(userId)
	c.Next()
}

func CheckLogin(c *gin.Context) {
	userId := c.GetUint64("userId")
	if userId == 0 {
		// 获取当前请求的完整URL作为重定向参数
		redirectURL := c.Request.URL.String()
		c.Redirect(http.StatusFound, "/login?redirect="+redirectURL)
		c.Abort()
		return
	}
	c.Next()
}

func CheckNeedLogin(c *gin.Context) {
	userId := c.GetUint64("userId")
	if userId != 0 {
		// 获取当前请求的完整URL作为重定向参数
		c.Redirect(http.StatusFound, `/`)
		c.Abort()
		return
	}
	c.Next()
}
