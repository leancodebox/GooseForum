package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	jwt "github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
)

const SkipUpdateUserActivity = "SkipUpdateUserActivity"

func JWTAuthCheck(c *gin.Context) {
	userId := JWTAuthGetUserId(c)
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailData("not authorized"))
		c.Abort()
		return
	}
	c.Set("userId", userId)
	c.Next()
	if !c.GetBool(SkipUpdateUserActivity) {
		eventbus.Publish(context.Background(), &eventhandlers.UserLastActiveUpdatedEvent{
			UserId:     userId,
			ActiveTime: time.Now(),
		})
	}
}

func JWTAuth(c *gin.Context) {
	userId := JWTAuthGetUserId(c)
	if userId != 0 {
		c.Set("userId", userId)
	}
	c.Next()
	if userId != 0 && !c.GetBool(SkipUpdateUserActivity) {
		eventbus.Publish(context.Background(), &eventhandlers.UserLastActiveUpdatedEvent{
			UserId:     userId,
			ActiveTime: time.Now(),
		})
	}
}

func JWTAuthGetUserId(c *gin.Context) uint64 {
	token := jwt.GetGinAccessToken(c)
	if token == "" {
		return 0
	}
	userId, newToken, err := jwt.VerifyTokenWithFresh(token)
	if err != nil {
		return 0
	}
	if token != newToken {
		jwt.TokenSetting(c, newToken)
	}
	return userId
}

func NoUpdateUserActivity(c *gin.Context) {
	c.Set(SkipUpdateUserActivity, true)
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
