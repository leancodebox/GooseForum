package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"net/http"
)

func CheckPermission(permissionType permission.Enum) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint64("userId")
		if userId == 0 {
			c.JSON(http.StatusUnauthorized, component.FailData("未登陆"))
			c.Abort()
			return
		}
		user, err := users.Get(userId)
		if err != nil {
			c.JSON(http.StatusForbidden, component.FailData("操作异常"))
			c.Abort()
			return
		}
		if permission.CheckRole(user.RoleId, permissionType) == false {
			msg := fmt.Sprintf("%s-不可操作-%s", user.Username, permissionType.Name())
			c.JSON(http.StatusForbidden, component.FailData(msg))
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckPermissionOrNoUser(permissionType permission.Enum) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint64("userId")
		if userId == 0 {
			return
		}
		user, err := users.Get(userId)
		if err != nil {
			c.JSON(http.StatusForbidden, component.FailData("操作异常"))
			c.Abort()
			return
		}
		if permission.CheckRole(user.RoleId, permissionType) == false {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
