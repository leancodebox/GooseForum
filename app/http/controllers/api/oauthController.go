package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/forum"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/oauthservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"github.com/markbates/goth/gothic"
)

// ProviderLogin 开始OAuth登录/绑定流程（根据登录状态自动判断）
func ProviderLogin(c *gin.Context) {
	provider := c.Param("provider")
	if !oauthservice.IsProviderEnabled(provider) {
		forum.RenderOAuthErrorPage(c, http.StatusNotFound, component.MessageOAuthCallbackFailed)
		return
	}
	q := c.Request.URL.Query()
	q.Set("provider", provider)
	c.Request.URL.RawQuery = q.Encode()
	if err := oauthservice.StartFlow(c.Writer, c.Request, provider, component.LoginUserId(c), c.Query("mode"), c.Query("redirect")); err != nil {
		forum.RenderOAuthErrorPage(c, http.StatusBadRequest, component.MessageOAuthCallbackFailed)
		return
	}
	authURL, err := gothic.GetAuthURL(c.Writer, c.Request)
	if err != nil {
		slog.Error("OAuth start failed", "provider", provider, "error", err)
		forum.RenderInternalOAuthErrorPage(c, component.MessageOAuthCallbackFailed)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// ProviderCallback 处理OAuth登录/绑定回调（根据登录状态自动判断）
func ProviderCallback(c *gin.Context) {
	provider := c.Param("provider")
	if !oauthservice.IsProviderEnabled(provider) {
		forum.RenderOAuthErrorPage(c, http.StatusNotFound, component.MessageOAuthCallbackFailed)
		return
	}
	q := c.Request.URL.Query()
	q.Set("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	// 完成 OAuth 流程
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		slog.Error("OAuth callback failed", "error", err)
		forum.RenderInternalOAuthErrorPage(c, component.MessageOAuthCallbackFailed)
		return
	}
	if gothUser.Provider != provider {
		slog.Warn("OAuth provider mismatch", "routeProvider", provider, "userProvider", gothUser.Provider)
		forum.RenderOAuthErrorPage(c, http.StatusBadRequest, component.MessageOAuthCallbackFailed)
		return
	}
	flow, err := oauthservice.ConsumeFlow(c.Writer, c.Request, provider)
	if err != nil {
		slog.Warn("OAuth flow validation failed", "provider", provider, "error", err)
		forum.RenderOAuthErrorPage(c, http.StatusBadRequest, component.MessageOAuthCallbackFailed)
		return
	}

	if flow.Mode == "bind" {
		currentUserId := component.LoginUserId(c)
		if currentUserId == 0 || currentUserId != flow.UserID {
			forum.RenderOAuthErrorPage(c, http.StatusUnauthorized, component.MessageAuthRequired)
			return
		}
		if user, ok := userservice.GetUserInfo(currentUserId); !ok || user.IsFrozen == users.StatusFrozen {
			forum.RenderOAuthErrorPage(c, http.StatusForbidden, component.MessagePermissionUserFrozen)
			return
		}

		// 绑定模式：处理OAuth绑定
		err = oauthservice.ProcessOAuthBind(currentUserId, gothUser)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/settings?tab=binding")
			return
		}
		// 绑定成功，重定向到账户设置页面
		c.Redirect(http.StatusTemporaryRedirect, "/settings?tab=binding")
	} else {
		// 登录模式：处理OAuth登录
		user, err := oauthservice.ProcessOAuthCallback(gothUser)
		if err != nil {
			slog.Error("Process OAuth callback failed", "error", err)
			forum.RenderInternalOAuthErrorPage(c, component.MessageOAuthProcessFailed)
			return
		}

		if user.IsFrozen == users.StatusFrozen {
			forum.RenderOAuthErrorPage(c, http.StatusForbidden, component.MessagePermissionUserFrozen)
			return
		}
		if hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification && user.IsActivated == users.ActivationPending {
			forum.RenderOAuthErrorPage(c, http.StatusForbidden, component.MessageAuthEmailUnverified)
			return
		}

		// 生成JWT token
		token, err := jwtopt.CreateNewTokenDefaultWithVersion(user.Id, user.TokenVersion)
		if err != nil {
			slog.Error("Generate JWT token failed", "error", err)
			forum.RenderInternalOAuthErrorPage(c, component.MessageOAuthTokenFailed)
			return
		}

		jwtopt.TokenSetting(c, token)
		redirect := flow.Redirect
		if redirect == "" {
			redirect = "/"
		}
		c.Redirect(http.StatusFound, redirect)
	}
}

// UnbindOAuth 解绑OAuth账户
func UnbindOAuth(req component.BetterRequest[component.Null]) component.Response {
	// 检查用户是否已登录
	userID := req.UserId

	provider := req.GinContext.Param("provider")

	// 解绑OAuth账户
	err := oauthservice.UnbindOAuth(userID, provider)
	if err != nil {
		return component.FailResponseCode(
			component.MessageOAuthUnbindFailed,

			component.MessageParams{"error": err.Error(), "provider": provider})

	}
	return component.SuccessResponseCode("解绑成功", component.MessageOAuthUnbindSuccess, component.MessageParams{"provider": provider})
}

// GetOAuthBindings 获取用户的OAuth绑定状态
func GetOAuthBindings(req component.BetterRequest[component.Null]) component.Response {
	return component.SuccessResponse(oauthservice.BindingProviders(req.UserId))
}
