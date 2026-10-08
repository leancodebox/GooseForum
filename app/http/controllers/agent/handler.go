package agent

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/models/forum/agenttokens"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/oidcproviderservice"
	"gorm.io/gorm"
)

type Handler struct {
	DB       func() *gorm.DB
	Provider func() (*core.Provider, error)
	Resolve  func(uint64) (accesscontrol.Snapshot, error)
	limits   limiter
}

type actor struct {
	User     users.EntityComplete
	ClientID string
	Source   string
	Scopes   []string
	Access   accesscontrol.Snapshot
	writable *bool
}

func New() *Handler {
	return &Handler{DB: dbconnect.Connect, Provider: oidcproviderservice.DefaultProvider, Resolve: accesscontrol.Resolve}
}

func (h *Handler) Register(engine *gin.Engine) {
	docs := engine.Group("/api/agent")
	docs.GET("/SKILL.md", h.skill)
	docs.GET("/v1/openapi.json", h.openapi)
	docs.GET("/v1/API.md", h.apiReference)
	docs.GET("/v1/site", h.site)
	g := docs.Group("/v1", h.authenticate)
	g.GET("/me", h.require(core.ScopeForumRead), h.me)
	g.GET("/categories", h.categories)
	g.GET("/topics", h.topics)
	g.GET("/search", h.search)
	g.GET("/topics/:topicId", h.topic)
	g.GET("/topics/:topicId/posts", h.posts)
	g.GET("/posts/:postId", h.post)
	g.POST("/topics", h.require(core.ScopeTopicsCreate), h.createTopic)
	g.POST("/topics/:topicId/posts", h.require(core.ScopePostsCreate), h.createPost)
	g.GET("/me/submissions", h.require(core.ScopeForumRead), h.submissions)
	g.GET("/me/submissions/:submissionId", h.require(core.ScopeForumRead), h.submission)
}

func requestID(c *gin.Context) string {
	if id := c.GetString("agentRequestID"); id != "" {
		return id
	}
	id := uuid.NewString()
	c.Set("agentRequestID", id)
	c.Header("X-Request-ID", id)
	return id
}

func fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message, "details": gin.H{}}, "requestId": requestID(c)})
}

func success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "requestId": requestID(c)})
}

func page(c *gin.Context, data any, pagination any) {
	c.JSON(http.StatusOK, gin.H{"data": data, "pagination": pagination, "requestId": requestID(c)})
}

func current(c *gin.Context) *actor { return c.MustGet("agentActor").(*actor) }

func enabled() bool { return hotdataserve.GetAgentSettingsConfigCache().Enabled }

func (h *Handler) authenticate(c *gin.Context) {
	requestID(c)
	config := hotdataserve.GetAgentSettingsConfigCache()
	c.Header("Cache-Control", "private, no-store")
	if !config.Enabled {
		fail(c, 503, "agent_api_disabled", "Agent API is disabled")
		return
	}
	if !h.limits.allow("all-ip:"+c.ClientIP(), config.IPPerMinute) {
		c.Header("Retry-After", "60")
		fail(c, 429, "rate_limited", "IP request limit reached")
		return
	}
	a := &actor{}
	header := c.GetHeader("Authorization")
	if header != "" {
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || len(parts[1]) > 512 {
			fail(c, 401, "invalid_token", "A valid Bearer credential is required")
			return
		}
		var userID, version uint64
		if strings.HasPrefix(parts[1], "gf_agent_") {
			if !config.ManualTokens {
				fail(c, 401, "invalid_token", "Manual tokens are disabled")
				return
			}
			var token agenttokens.Entity
			err := h.DB().Where("hash = ?", agenttokens.Digest(parts[1])).Take(&token).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				fail(c, 503, "auth_unavailable", "Credential validation is unavailable")
				return
			}
			if err != nil || token.RevokedAt != nil || !time.Now().Before(token.ExpiresAt) || !agenttokens.ValidScopes(token.Scopes) {
				fail(c, 401, "invalid_token", "Credential is expired, revoked or invalid")
				return
			}
			userID, version = token.UserID, token.UserVersion
			a.Source, a.Scopes = agenttokens.Digest("manual:"+token.ID), token.Scopes
		} else {
			provider, err := h.Provider()
			if err != nil {
				fail(c, 503, "auth_unavailable", "Browser authorization is unavailable")
				return
			}
			token, err := provider.ValidateAccessToken(c.Request.Context(), parts[1])
			if err != nil {
				if errors.Is(err, core.ErrServer) {
					fail(c, 503, "auth_unavailable", "Credential validation is unavailable")
				} else {
					fail(c, 401, "invalid_token", "Credential is expired, revoked or invalid")
				}
				return
			}
			if !slices.Contains(token.Scopes, core.ScopeForumRead) {
				fail(c, 403, "insufficient_scope", "forum:read is required")
				return
			}
			userID, err = strconv.ParseUint(token.UserID, 10, 64)
			if err != nil || userID == 0 {
				fail(c, 401, "invalid_token", "Invalid account")
				return
			}
			version = token.UserVersion
			a.ClientID, a.Scopes = token.ClientID, token.Scopes
			// Stable across token refresh, and unambiguous across user/client pairs.
			identity, _ := json.Marshal([]string{token.UserID, token.ClientID})
			a.Source = agenttokens.Digest("oauth:" + string(identity))
		}
		err := h.DB().Where("id = ?", userID).Take(&a.User).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, 503, "auth_unavailable", "Account lookup is unavailable")
			return
		}
		if err != nil || a.User.TokenVersion != version || a.User.EffectiveRestriction(time.Now()) == users.RestrictionBanned {
			fail(c, 401, "invalid_token", "Account credentials changed or account is unavailable")
			return
		}
	}
	key, maximum := "ip:"+c.ClientIP(), config.AnonymousReadPerMinute
	if a.User.Id != 0 {
		key, maximum = "source:"+a.Source, config.ReadPerMinute
	}
	if !h.limits.allow(key, maximum) {
		c.Header("Retry-After", "60")
		fail(c, 429, "rate_limited", "Request limit reached")
		return
	}
	if a.User.Id != 0 && !h.limits.allow("user:"+strconv.FormatUint(a.User.Id, 10), config.UserReadPerMinute) {
		c.Header("Retry-After", "60")
		fail(c, 429, "rate_limited", "Account request limit reached")
		return
	}
	if c.Request.Method == http.MethodPost && a.User.Id != 0 && !h.limits.allow("write:"+strconv.FormatUint(a.User.Id, 10), config.WritePerMinute) {
		c.Header("Retry-After", "60")
		fail(c, 429, "rate_limited", "Write limit reached")
		return
	}
	var err error
	a.Access, err = h.Resolve(a.User.Id)
	if err != nil {
		fail(c, 503, "permissions_unavailable", "Permissions are unavailable")
		return
	}
	c.Set("agentActor", a)
	started := time.Now()
	c.Next()
	slog.Info("agent request", "requestId", requestID(c), "credentialId", a.Source, "clientId", a.ClientID, "userId", a.User.Id, "operation", c.FullPath(), "status", c.Writer.Status(), "duration", time.Since(started))
}

func (h *Handler) require(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		a := current(c)
		if a.User.Id == 0 {
			c.Header("WWW-Authenticate", `Bearer realm="GooseForum Agent API"`)
			fail(c, 401, "auth_required", "Browser authorization or a manual token is required")
			return
		}
		if !slices.Contains(a.Scopes, scope) {
			fail(c, 403, "insufficient_scope", "Required scope: "+scope)
			return
		}
		c.Next()
	}
}

func decode(c *gin.Context, dest any) bool {
	if !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
		fail(c, 400, "invalid_request", "Content-Type must be application/json")
		return false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	err := d.Decode(dest)
	if err == nil {
		var extra any
		err = d.Decode(&extra)
		if err == io.EOF {
			err = nil
		} else if err == nil {
			err = errors.New("unexpected trailing JSON")
		}
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			fail(c, 413, "request_too_large", "Request exceeds 1 MiB")
		} else {
			fail(c, 400, "invalid_request", "Invalid JSON or unknown fields")
		}
		return false
	}
	return true
}

func id(value string) (uint64, bool) {
	v, err := strconv.ParseUint(value, 10, 64)
	return v, err == nil && v > 0 && strconv.FormatUint(v, 10) == value
}

func limit(c *gin.Context) (int, bool) {
	if c.Query("limit") == "" {
		return 20, true
	}
	n, err := strconv.Atoi(c.Query("limit"))
	if err != nil || n < 1 || n > 50 {
		fail(c, 400, "invalid_request", "limit must be between 1 and 50")
		return 0, false
	}
	return n, true
}

type bucket struct {
	start time.Time
	count int
}
type limiter struct {
	mu      sync.Mutex
	entries map[string]bucket
	cleaned time.Time
}

func (l *limiter) allow(key string, max int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if max < 1 {
		max = 1
	}
	now := time.Now()
	if l.entries == nil {
		l.entries = map[string]bucket{}
	}
	if now.Sub(l.cleaned) >= time.Minute {
		for key, b := range l.entries {
			if now.Sub(b.start) >= time.Minute {
				delete(l.entries, key)
			}
		}
		l.cleaned = now
	}
	b, exists := l.entries[key]
	if !exists && len(l.entries) >= 10000 {
		return false
	}
	if now.Sub(b.start) >= time.Minute {
		b = bucket{start: now}
	}
	if b.count >= max {
		return false
	}
	b.count++
	l.entries[key] = b
	return true
}
