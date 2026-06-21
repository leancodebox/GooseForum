package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/forum"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

func CheckPermission(permissionType permission.Enum) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleId, ok := resolveRoleId(c)
		if !ok {
			return
		}

		if !permission.CheckRole(roleId, permissionType) {
			c.JSON(http.StatusForbidden, component.FailDataCode(
				component.MessagePermissionDenied,
				component.MessageParams{"permission": permissionType.Name()}))
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckAnyPermission(c *gin.Context) {
	roleId, ok := resolveRoleId(c)
	if !ok {
		return
	}
	if !permission.CheckAnyRole(roleId) {
		c.JSON(http.StatusForbidden, component.FailDataCode(component.MessagePermissionDenied, nil))
		c.Abort()
		return
	}
	c.Next()
}

func CheckAnyPermissionOrNotFound(c *gin.Context) {
	roleId, ok := resolveRoleId(c)
	if !ok {
		return
	}
	if !permission.CheckAnyRole(roleId) {
		forum.RenderNotFoundPage(c, component.MessagePageNotFound)
		c.Abort()
		return
	}
	c.Next()
}

func CheckWritableAccount(c *gin.Context) {
	userId := c.GetUint64("userId")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailDataCode(component.MessageAuthRequired, nil))
		c.Abort()
		return
	}

	user, ok := userservice.GetUserInfo(userId)
	if !ok {
		c.JSON(http.StatusForbidden, component.FailDataCode(component.MessagePermissionResolveFailed, nil))
		c.Abort()
		return
	}
	if user.IsFrozen == users.StatusFrozen {
		c.JSON(http.StatusForbidden, component.FailDataCode(
			component.MessagePermissionUserFrozen,
			component.MessageParams{
				"action":     "写入",
				"actionCode": string(component.PermissionActionWrite),
			},
		))
		c.Abort()
		return
	}
	c.Next()
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

		if !permission.CheckRole(roleId, permissionType) {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

func resolveRoleId(c *gin.Context) (uint64, bool) {
	userId := c.GetUint64("userId")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailDataCode(component.MessageAuthRequired, nil))
		c.Abort()
		return 0, false
	}

	if val, exists := c.Get("roleId"); exists {
		if roleId, ok := val.(uint64); ok && roleId != 0 {
			return roleId, true
		}
	}

	roleId, ok := userservice.GetUserRoleId(userId)
	if !ok {
		c.JSON(http.StatusForbidden, component.FailDataCode(component.MessagePermissionResolveFailed, nil))
		c.Abort()
		return 0, false
	}
	return roleId, true
}
