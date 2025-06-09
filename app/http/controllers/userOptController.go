package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
)

// 添加激活处理函数
func ActivateAccount(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		renderActivationPage(c, false, "无效的激活链接")
		return
	}

	// 解析激活令牌
	claims, err := tokenservice.ParseActivationToken(token)
	if err != nil {
		renderActivationPage(c, false, "激活链接已过期或无效")
		return
	}

	// 获取用户信息
	user, err := users.Get(claims.UserId)
	if err != nil {
		renderActivationPage(c, false, "用户不存在")
		return
	}

	// 检查邮箱是否匹配
	if user.Email != claims.Email {
		renderActivationPage(c, false, "激活链接无效")
		return
	}

	// 激活账号
	err = user.Activate()
	if err != nil {
		renderActivationPage(c, false, "激活失败")
		return
	}

	renderActivationPage(c, true, "账号激活成功")
}

// 添加新的辅助函数
func renderActivationPage(c *gin.Context, success bool, message string) {
	status := "失败"
	if success {
		status = "成功"
	}
	viewrender.Render(c, "activate.gohtml", map[string]any{
		"Status":  status,
		"Message": message,
		"Success": success,
	})
}
