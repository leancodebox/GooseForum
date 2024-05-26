package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/spf13/cast"
	"net/http"
)

func userLimit(c *gin.Context) {
	userIdData, _ := c.Get("userId")
	userId := cast.ToUint64(userIdData)
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailData("未登陆"))
		c.Abort()
	}
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailData("未登陆"))
		c.Abort()
	}
	//user, err := users.Get(userId)

	c.Next()
}
