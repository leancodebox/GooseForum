package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

// 获取用户，如果没有初始化就获取不到
func getUserLazy(c *gin.Context) (users.EntityComplete, bool) {
	if user, ok := c.Get("user"); ok {
		if userEntity, uOk := user.(users.EntityComplete); uOk {
			return userEntity, true
		}
	}
	return users.EntityComplete{}, false
}

// 获取用户
func getUser(c *gin.Context) (users.EntityComplete, bool) {
	if user, ok := c.Get("user"); ok {
		if userEntity, uOk := user.(users.EntityComplete); uOk {
			return userEntity, true
		}
	}
	userId := c.GetUint64("userId")
	if userId == 0 {
		return users.EntityComplete{}, false
	}
	user, err := users.Get(userId)
	if err != nil {
		return users.EntityComplete{}, false
	}
	c.Set("user", user)
	return user, false
}
