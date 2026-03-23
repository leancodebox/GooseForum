package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
)

func CheckPermission(permissionType permission.Enum) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint64("userId")
		if userId == 0 {
			c.JSON(http.StatusUnauthorized, component.FailData("未登陆"))
			c.Abort()
			return
		}

		var roleId uint64
		// 尝试从 Context 获取 RoleId
		if val, exists := c.Get("roleId"); exists {
			roleId = val.(uint64)
		}

		// 如果 roleId 为 0，回退到查库
		if roleId == 0 {
			var err error
			roleId, err = users.GetRoleId(userId)
			if err != nil {
				c.JSON(http.StatusForbidden, component.FailData("操作异常"))
				c.Abort()
				return
			}
		}

		if permission.CheckRole(roleId, permissionType) == false {
			msg := fmt.Sprintf("User(%d)-不可操作-%s", userId, permissionType.Name())
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

		var roleId uint64
		if val, exists := c.Get("roleId"); exists {
			roleId = val.(uint64)
		}

		if roleId == 0 {
			var err error
			roleId, err = users.GetRoleId(userId)
			if err != nil {
				c.JSON(http.StatusForbidden, component.FailData("操作异常"))
				c.Abort()
				return
			}
		}

		if permission.CheckRole(roleId, permissionType) == false {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
