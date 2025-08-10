package controllers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/service/oauthservice"
	"github.com/markbates/goth/gothic"
)

// ProviderLogin 开始GitHub OAuth登录流程
func ProviderLogin(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()
	// 开始OAuth流程
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// ProviderCallback 处理GitHub OAuth回调
func ProviderCallback(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()
	// 完成OAuth流程
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		slog.Error("GitHub OAuth callback failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "OAuth认证失败",
		})
		return
	}

	// 处理OAuth回调
	user, _, err := oauthservice.ProcessOAuthCallback(gothUser)
	if err != nil {
		slog.Error("Process OAuth callback failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理OAuth回调失败",
		})
		return
	}

	// 生成JWT token
	token, err := jwtopt.CreateNewTokenDefault(user.Id)
	if err != nil {
		slog.Error("Generate JWT token failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成token失败",
		})
		return
	}

	jwtopt.TokenSetting(c, token)

	c.Redirect(http.StatusFound, "/")
}

// ProviderLogout 处理OAuth提供商登出
func ProviderLogout(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()
	// 执行登出
	err := gothic.Logout(c.Writer, c.Request)
	if err != nil {
		slog.Error("GitHub OAuth logout failed", "error", err)
	}

	// 清除cookie
	c.SetCookie("token", "", -1, "/", "", false, true)

	c.Redirect(http.StatusFound, "/")
}
