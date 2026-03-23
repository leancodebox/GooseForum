package controllers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/oauthservice"
	"github.com/markbates/goth/gothic"
)

// ProviderLogin 开始OAuth登录/绑定流程（根据登录状态自动判断）
func ProviderLogin(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()
	// 开始 OAuth 流程
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// ProviderCallback 处理OAuth登录/绑定回调（根据登录状态自动判断）
func ProviderCallback(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()

	// 完成 OAuth 流程
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		slog.Error("OAuth callback failed", "error", err)
		// 如果 gothic 已经写入了响应，就不要再写了
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "OAuth 认证失败",
		})
		return
	}

	// 检查是否为绑定模式（用户已登录）
	currentUserInfo := component.GetLoginUser(c)
	currentUserId := currentUserInfo.UserId

	if currentUserId > 0 {
		// 绑定模式：处理OAuth绑定
		err = oauthservice.ProcessOAuthBind(currentUserId, gothUser)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/settings?setting-tab=account&error="+err.Error())
			return
		}
		// 绑定成功，重定向到账户设置页面
		c.Redirect(http.StatusTemporaryRedirect, "/settings?setting-tab=account&success=bind_success")
	} else {
		// 登录模式：处理OAuth登录
		user, err := oauthservice.ProcessOAuthCallback(gothUser)
		if err != nil {
			slog.Error("Process OAuth callback failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "处理 OAuth 回调失败",
			})
			return
		}

		// 检查用户状态
		if user.IsFrozen == users.StatusFrozen {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "您的账号已被禁用，如有疑问请联系管理员",
			})
			return
		}

		if user.IsActivated == users.ActivationPending {
			user.IsActivated = users.ActivationSuccess
			// 更新用户状态
			err = users.Save(user)
			if err != nil {
				slog.Error("Update user activation status failed", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "更新用户激活状态失败",
				})
				return
			}
		}

		// 生成JWT token
		token, err := jwtopt.CreateNewTokenDefault(user.Id)
		if err != nil {
			slog.Error("Generate JWT token failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "生成 token 失败",
			})
			return
		}

		jwtopt.TokenSetting(c, token)
		c.Redirect(http.StatusFound, "/")
	}
}

// UnbindOAuth 解绑OAuth账户
func UnbindOAuth(req component.BetterRequest[null]) component.Response {
	// 检查用户是否已登录
	userID := req.UserId

	provider := req.GinContext.Param("provider")

	// 解绑OAuth账户
	err := oauthservice.UnbindOAuth(userID, provider)
	if err != nil {
		return component.FailResponse(err.Error())
	}
	return component.SuccessResponse("解绑成功")
}

// GetOAuthBindings 获取用户的OAuth绑定状态
func GetOAuthBindings(req component.BetterRequest[null]) component.Response {
	// 检查用户是否已登录
	userID := req.UserId

	// 获取用户的 OAuth 绑定
	bindings := oauthservice.GetUserOAuthBindings(userID)

	// 构建响应数据
	result := make(map[string]any)
	for provider, oauth := range bindings {
		result[provider] = map[string]any{
			"bound":     true,
			"provider":  oauth.Provider,
			"createdAt": oauth.CreatedAt,
			"updatedAt": oauth.UpdatedAt,
		}
	}

	// 添加未绑定的提供商
	allProviders := []string{"github", "google"}
	for _, provider := range allProviders {
		if _, exists := result[provider]; !exists {
			result[provider] = map[string]any{
				"bound": false,
			}
		}
	}
	return component.SuccessResponse(result)

}
