package api

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/forum"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/leancodebox/GooseForum/app/service/loginlogservice"
	"github.com/leancodebox/GooseForum/app/service/mfaservice"
	"github.com/leancodebox/GooseForum/app/service/oauthservice"
	"github.com/markbates/goth/gothic"
)

// ProviderLogin 开始OAuth登录/绑定流程（根据登录状态自动判断）
func ProviderLogin(c *gin.Context) {
	provider := c.Param("provider")
	if !oauthservice.IsProviderEnabled(provider) {
		slog.Warn("OAuth start provider unavailable", "provider", provider)
		oauthStartFailure(c, http.StatusNotFound, component.MessageOAuthCallbackFailed)
		return
	}
	if c.Request.Method == http.MethodPost && c.Query("mode") != "bind" {
		oauthStartFailure(c, http.StatusBadRequest, component.MessageRequestInvalidParams)
		return
	}
	var authority oauthservice.BindingAuthority
	if c.Query("mode") == "bind" {
		var code component.MessageCode
		authority, code = authorizeOAuthBinding(c)
		if code != "" {
			oauthStartFailure(c, http.StatusForbidden, code)
			return
		}
	}
	q := c.Request.URL.Query()
	q.Set("provider", provider)
	c.Request.URL.RawQuery = q.Encode()
	if err := oauthservice.StartFlow(c.Writer, c.Request, provider, component.LoginUserId(c), c.Query("mode"), c.Query("redirect"), authority); err != nil {
		slog.Error("OAuth start flow failed", "provider", provider, "error", err)
		oauthStartFailure(c, http.StatusBadRequest, component.MessageOAuthCallbackFailed)
		return
	}
	authURL, err := gothic.GetAuthURL(c.Writer, c.Request)
	if err != nil {
		slog.Error("OAuth start failed", "provider", provider, "error", err)
		oauthStartFailure(c, http.StatusInternalServerError, component.MessageOAuthCallbackFailed)
		return
	}
	if c.Request.Method == http.MethodPost {
		c.JSON(http.StatusOK, component.SuccessData(map[string]any{"redirect": authURL}))
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func oauthStartFailure(c *gin.Context, status int, code component.MessageCode) {
	if c.Request.Method == http.MethodPost {
		c.JSON(status, component.FailDataCode(code, nil))
		return
	}
	forum.RenderOAuthErrorPage(c, status, code)
}

func authorizeOAuthBinding(c *gin.Context) (oauthservice.BindingAuthority, component.MessageCode) {
	userID, sessionID := component.LoginUserId(c), c.GetUint64("sessionId")
	if userID == 0 || sessionID == 0 {
		return oauthservice.BindingAuthority{}, component.MessageAuthRequired
	}
	user, err := users.Get(userID)
	if err != nil || user.EffectiveRestriction(time.Now()) == users.RestrictionBanned {
		return oauthservice.BindingAuthority{}, component.MessageAuthRequired
	}
	if user.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) {
		return oauthservice.BindingAuthority{}, component.MessageAuthEmailUnverified
	}
	var req UnbindOAuthReq
	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBindJSON(&req); err != nil || len(req.MFACode) > 128 {
			return oauthservice.BindingAuthority{}, component.MessageRequestInvalidParams
		}
	}
	// GET carries no code: an MFA account must start binding through POST.
	if err := mfaservice.VerifySecondFactor(c, userID, req.MFACode); err != nil {
		return oauthservice.BindingAuthority{}, mfaErrorCode(err)
	}
	return oauthservice.BindingAuthority{SessionID: sessionID, TokenVersion: user.TokenVersion}, ""
}

// ProviderCallback 处理OAuth登录/绑定回调（根据登录状态自动判断）
func ProviderCallback(c *gin.Context) {
	provider := c.Param("provider")
	completed := false
	defer func() {
		if !completed {
			loginlogservice.Failure(c, "oauth", provider, "callback_rejected")
		}
	}()
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
		slog.Error("OAuth callback failed", "provider", provider, "error", err)
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
		user, err := users.Get(currentUserId)
		if err != nil || user.EffectiveRestriction(time.Now()) == users.RestrictionBanned || user.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) {
			forum.RenderOAuthErrorPage(c, http.StatusForbidden, component.MessagePermissionUserFrozen)
			return
		}
		if err := flow.ValidateBinding(currentUserId, c.GetUint64("sessionId"), user.TokenVersion); err != nil {
			forum.RenderOAuthErrorPage(c, http.StatusUnauthorized, component.MessageAuthRequired)
			return
		}

		// 绑定模式：处理OAuth绑定
		err = oauthservice.ProcessOAuthBind(currentUserId, gothUser)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/settings?tab=binding")
			return
		}
		// 绑定成功，重定向到账户设置页面
		completed = true
		loginlogservice.Record(c, loginlogservice.Event{UserID: currentUserId, Action: "oauth_bind", Method: "oauth", Provider: provider, Result: "success"})
		c.Redirect(http.StatusTemporaryRedirect, "/settings?tab=binding")
	} else {
		// 登录模式：处理OAuth登录
		user, err := oauthservice.ProcessOAuthCallback(gothUser, c.ClientIP())
		if err != nil {
			slog.Error("Process OAuth callback failed", "error", err)
			mapped := registrationError(err).(component.MessageError)
			if mapped.Code == component.MessageAuthRegisterFailed {
				forum.RenderInternalOAuthErrorPage(c, component.MessageOAuthProcessFailed)
			} else {
				forum.RenderOAuthErrorPage(c, http.StatusForbidden, mapped.Code)
			}
			return
		}

		if user.EffectiveRestriction(time.Now()) == users.RestrictionBanned {
			forum.RenderOAuthErrorPage(c, http.StatusForbidden, component.MessagePermissionUserFrozen)
			return
		}
		if user.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) {
			forum.RenderOAuthErrorPage(c, http.StatusForbidden, component.MessageAuthEmailUnverified)
			return
		}

		challenge, err := mfaservice.CompleteFirstFactor(c, user.Id, user.TokenVersion, authsessionservice.LoginDetails{Method: "oauth", Provider: provider}, flow.Redirect)
		if err != nil {
			slog.Error("Generate JWT token failed", "error", err)
			code := component.MessageOAuthTokenFailed
			if errors.Is(err, mfaservice.ErrUnavailable) || errors.Is(err, mfaservice.ErrVerification) {
				code = mfaErrorCode(err)
			}
			forum.RenderInternalOAuthErrorPage(c, code)
			return
		}
		if challenge {
			completed = true
			c.Redirect(http.StatusFound, "/login?mfa=1")
			return
		}
		completed = true
		redirect := flow.Redirect
		if redirect == "" {
			redirect = "/"
		}
		c.Redirect(http.StatusFound, redirect)
	}
}

// UnbindOAuth 解绑OAuth账户
type UnbindOAuthReq struct {
	MFACode string `json:"mfaCode"`
}

func UnbindOAuth(req component.BetterRequest[UnbindOAuthReq]) component.Response {
	// 检查用户是否已登录
	userID := req.UserId

	provider := req.GinContext.Param("provider")
	if err := mfaservice.VerifySecondFactor(req.GinContext, userID, req.Params.MFACode); err != nil {
		return component.FailResponseCode(mfaErrorCode(err), nil)
	}

	// 解绑OAuth账户
	err := oauthservice.UnbindOAuth(userID, provider)
	if err != nil {
		return component.FailResponseCode(
			component.MessageOAuthUnbindFailed,

			component.MessageParams{"error": err.Error(), "provider": provider})

	}
	loginlogservice.Record(req.GinContext, loginlogservice.Event{UserID: userID, Action: "oauth_unbind", Method: "oauth", Provider: provider, Result: "success"})
	return component.SuccessResponseCode("解绑成功", component.MessageOAuthUnbindSuccess, component.MessageParams{"provider": provider})
}

// GetOAuthBindings 获取用户的OAuth绑定状态
func GetOAuthBindings(req component.BetterRequest[component.Null]) component.Response {
	return component.SuccessResponse(oauthservice.BindingProviders(req.UserId))
}
