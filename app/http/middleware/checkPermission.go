package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

func CheckPermission(permissionType permission.Enum) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint64("userId")
		if userId == 0 {
			c.JSON(http.StatusUnauthorized, component.FailDataCode(component.MessageAuthRequired, nil))
			c.Abort()
			return
		}

		var roleId uint64
		// 尝试从 Context 获取 RoleId
		if val, exists := c.Get("roleId"); exists {
			roleId = val.(uint64)
		}

		// 如果 roleId 为 0，回退到用户信息缓存
		if roleId == 0 {
			var ok bool
			roleId, ok = userservice.GetUserRoleId(userId)
			if !ok {
				c.JSON(http.StatusForbidden, component.FailDataCode(component.MessagePermissionResolveFailed, nil))
				c.Abort()
				return
			}
		}

		if permission.CheckRole(roleId, permissionType) == false {
			c.JSON(http.StatusForbidden, component.FailDataCode(
				component.MessagePermissionDenied,
				component.MessageParams{"permission": permissionType.Name()}))
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
			var ok bool
			roleId, ok = userservice.GetUserRoleId(userId)
			if !ok {
				c.JSON(http.StatusForbidden, component.FailDataCode(component.MessagePermissionResolveFailed, nil))
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
