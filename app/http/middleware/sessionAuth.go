package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/trustlevelservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

const SkipUpdateUserActivity = "SkipUpdateUserActivity"

func SessionAuthCheck(c *gin.Context) {
	userId := SessionAuthGetUserId(c)
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailDataCode(component.MessageAuthRequired, nil))
		c.Abort()
		return
	}
	c.Set("userId", userId)
	reviewTrust(c, userId)
	c.Next()
	if !c.GetBool(SkipUpdateUserActivity) {
		eventbus.Publish(context.Background(), &eventhandlers.UserLastActiveUpdatedEvent{
			UserId:     userId,
			ActiveTime: time.Now(),
		})
	}
}

func SessionAuth(c *gin.Context) {
	userId := SessionAuthGetUserId(c)
	if userId != 0 {
		c.Set("userId", userId)
		reviewTrust(c, userId)
	}
	c.Next()
	if userId != 0 && !c.GetBool(SkipUpdateUserActivity) {
		eventbus.Publish(context.Background(), &eventhandlers.UserLastActiveUpdatedEvent{
			UserId:     userId,
			ActiveTime: time.Now(),
		})
	}
}

// SessionAuthSilent resolves an optional login without emitting activity events.
// It is intended for high-frequency protected resources such as inline images.
func SessionAuthSilent(c *gin.Context) {
	if userID := SessionAuthGetUserId(c); userID != 0 {
		c.Set("userId", userID)
	}
	c.Next()
}

func SessionAuthGetUserId(c *gin.Context) uint64 {
	token, fromCookie := authsessionservice.AccessToken(c)
	if token == "" {
		return 0
	}
	auth, err := authsessionservice.Authenticate(c, token, fromCookie)
	if err != nil {
		return 0
	}
	_, ok := userservice.GetUserInfo(auth.Session.UserId)
	if !ok {
		return 0
	}
	c.Set("sessionId", auth.Session.Id)
	c.Set("authenticatedAccount", auth.Account)
	c.Set("authTime", auth.Session.AuthTime.Unix())
	// Old and federated sessions do not prove fresh credential verification.
	if auth.Session.Reauthenticated {
		c.Set("authId", auth.Session.AuthId)
	}
	return auth.Session.UserId
}

func NoUpdateUserActivity(c *gin.Context) {
	c.Set(SkipUpdateUserActivity, true)
	c.Next()
}

func reviewTrust(c *gin.Context, id uint64) {
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/p/post/") || strings.HasPrefix(path, "/file/") || strings.HasPrefix(path, "/api/forum/unread-status") || strings.HasPrefix(path, "/api/forum/notifications") || strings.HasPrefix(path, "/api/forum/posts/window") || strings.HasPrefix(path, "/api/forum/chat/") {
		return
	}
	if c.GetBool("trustReviewed") || strings.Contains(c.Request.URL.Path, "/admin/") || c.Request.Method == "HEAD" || strings.Contains(c.GetHeader("Purpose"), "prefetch") || strings.Contains(c.GetHeader("Sec-Purpose"), "prefetch") {
		return
	}
	c.Set("trustReviewed", true)
	state, ok := c.Get("authenticatedAccount")
	account, valid := state.(users.AccountState)
	if !ok || !valid || account.Id != id {
		return
	}
	if err := trustlevelservice.ReviewAuthenticated(account, time.Now()); err != nil {
		slog.Warn("review trust failed", "userId", id, "error", err)
	}
}

func CheckLogin(c *gin.Context) {
	userId := c.GetUint64("userId")
	if userId == 0 {
		// 获取当前请求的完整URL作为重定向参数
		redirectURL := c.Request.URL.String()
		c.Redirect(http.StatusFound, "/login?redirect="+redirectURL)
		c.Abort()
		return
	}
	c.Next()
}
