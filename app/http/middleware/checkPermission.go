package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/spf13/cast"
	"net/http"
)

func CheckPermission(permissionType permission.Enum) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		user, err := users.Get(userId)
		if err != nil {
			c.JSON(http.StatusForbidden, component.FailData("操作异常"))
			c.Abort()
		}
		if permission.CheckUser(user.Id, permissionType) == false {
			msg := fmt.Sprintf("%s-不可操作-%s", user.Username, permissionType.Name())
			c.JSON(http.StatusForbidden, component.FailData(msg))
			c.Abort()
		}
	}
}
