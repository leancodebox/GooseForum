package authsessionservice

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/loginlogservice"
	"gorm.io/gorm"
)

const (
	maxSessions  = 20
	seenInterval = 15 * time.Minute
	renewWindow  = 24 * time.Hour
)

var ErrInvalidSession = errors.New("invalid session")

type LoginDetails struct {
	Method          string
	Provider        string
	Reauthenticated bool
	AuthTime        time.Time
}

type Authenticated struct {
	Session authsessions.Token
	Renewed bool
}

func lifetime() time.Duration {
	seconds := preferences.GetInt64("authsession.validTime", preferences.GetInt64("jwtopt.validTime", 86400*7))
	if seconds < 86400 {
		seconds = 86400
	}
	return time.Duration(seconds) * time.Second
}

func renewalWindow() time.Duration {
	return min(renewWindow, lifetime()/4)
}

func hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func Issue(c *gin.Context, userID, tokenVersion uint64, details LoginDetails) error {
	return issue(c, userID, tokenVersion, details, nil)
}

// IssueVerified consumes the second factor before creating the session.
func IssueVerified(c *gin.Context, userID, tokenVersion uint64, details LoginDetails, verify func() error) error {
	return issue(c, userID, tokenVersion, details, verify)
}

func issue(c *gin.Context, userID, tokenVersion uint64, details LoginDetails, verify func() error) error {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return err
	}
	raw := base64.RawURLEncoding.EncodeToString(secret)
	clear(secret)
	now := time.Now()
	authTime := details.AuthTime
	if authTime.IsZero() {
		authTime = now
	}
	session := authsessions.Token{
		UserId: userID, TokenHash: hash(raw), TokenVersion: tokenVersion,
		AuthMethod: details.Method, OAuthProvider: details.Provider,
		AuthTime: authTime, Reauthenticated: details.Reauthenticated,
		ClientIP: c.ClientIP(), UserAgent: truncate(c.Request.UserAgent(), 512),
		CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(lifetime()),
	}
	if details.Reauthenticated {
		var err error
		session.AuthId, err = rawAuthID()
		if err != nil {
			return err
		}
	}
	state, stateErr := users.GetAccountState(userID)
	if stateErr != nil || state.TokenVersion != tokenVersion || state.EffectiveRestriction(now) == users.RestrictionBanned || state.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) {
		return ErrInvalidSession
	}
	if verify != nil {
		if err := verify(); err != nil {
			return err
		}
	}
	// Verification stays consumed if session creation fails; retry requires another code.
	if err := authsessions.CreateToken(&session); err != nil {
		return err
	}
	logEvent(c, session, "login_success")
	// The limit is a convenience bound; validity always comes from the row lookup.
	excess, _ := authsessions.ExcessTokens(userID, now, maxSessions)
	for _, old := range excess {
		if revokeID(old.Id, userID, now) {
			cachedSessions.InvalidateUser(old.UserId)
			logEvent(c, old, "limit")
		}
	}
	SetCookie(c, raw, lifetime())
	return nil
}

func InvalidateUser(userID uint64) { cachedSessions.InvalidateUser(userID) }

func LogSecurityEvent(c *gin.Context, userID uint64, action string) {
	method := ""
	if strings.HasPrefix(action, "mfa_") {
		method = "mfa"
	}
	session := authsessions.Token{UserId: userID, AuthMethod: method}
	if c != nil {
		session.Id = c.GetUint64("sessionId")
	}
	logEvent(c, session, action)
}

func rawAuthID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	defer clear(b)
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func AccessToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer "), false
	}
	token, _ := c.Cookie("access_token")
	return token, true
}

func Authenticate(c *gin.Context, raw string, fromCookie bool) (Authenticated, error) {
	if len(raw) != 43 {
		return Authenticated{}, ErrInvalidSession
	}
	key := hash(raw)
	now := time.Now()
	session, ok := cachedSessions.Get(key, now)
	if !ok {
		var err error
		session, err = loadSession(key, now)
		if err != nil {
			return Authenticated{}, err
		}
	}
	state, stateErr := users.GetAccountState(session.UserId)
	if stateErr != nil || state.TokenVersion != session.TokenVersion || state.EffectiveRestriction(now) == users.RestrictionBanned || state.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) {
		cachedSessions.InvalidateUser(session.UserId)
		return Authenticated{}, ErrInvalidSession
	}
	result := Authenticated{Session: session}
	needsRenewal := session.ExpiresAt.Sub(now) < renewalWindow()
	if needsRenewal {
		newExpiry := now.Add(lifetime())
		updated, err := authsessions.RenewToken(session.Id, now, newExpiry)
		if err != nil {
			return Authenticated{}, err
		}
		if updated {
			result.Renewed = true
			session.ExpiresAt = newExpiry
			cachedSessions.UpdateIfPresent(key, session)
		} else {
			cachedSessions.InvalidateUser(session.UserId)
		}
	}
	if session.LastSeenAt.Before(now.Add(-seenInterval)) {
		updated, err := authsessions.TouchToken(session.Id, now.Add(-seenInterval), now)
		if err == nil && updated {
			session.LastSeenAt = now
			cachedSessions.UpdateIfPresent(key, session)
		}
	}
	if fromCookie && needsRenewal {
		SetCookie(c, raw, lifetime())
	}
	return result, nil
}

func loadSession(key string, now time.Time) (authsessions.Token, error) {
	for range 2 {
		generation := cachedSessions.Generation()
		session, err := authsessions.GetValidToken(key, now)
		if err != nil {
			return authsessions.Token{}, ErrInvalidSession
		}
		if cachedSessions.Store(key, session, now, generation) {
			return session, nil
		}
	}
	return authsessions.Token{}, ErrInvalidSession
}

func SetCookie(c *gin.Context, raw string, ttl time.Duration) {
	c.Header("New-Token", raw)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", raw, int(ttl/time.Second), "/", "", !setting.IsLocal(), true)
}

func ClearCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", !setting.IsLocal(), true)
}

func RevokeRaw(c *gin.Context, raw string) error {
	if len(raw) != 43 {
		return nil
	}
	session, err := authsessions.GetTokenByHash(hash(raw))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	revoked, err := authsessions.RevokeToken(session.UserId, session.Id, time.Now())
	if err != nil {
		return err
	}
	if revoked {
		cachedSessions.InvalidateUser(session.UserId)
		logEvent(c, session, "logout")
	}
	return nil
}

func revokeID(id, userID uint64, now time.Time) bool {
	revoked, err := authsessions.RevokeToken(userID, id, now)
	return err == nil && revoked
}

func Revoke(c *gin.Context, userID, id, currentID uint64) error {
	if id == currentID {
		return ErrInvalidSession
	}
	session, err := authsessions.GetUserToken(userID, id)
	if err != nil {
		return ErrInvalidSession
	}
	if !revokeID(id, userID, time.Now()) {
		return ErrInvalidSession
	}
	cachedSessions.InvalidateUser(userID)
	logEvent(c, session, "revoke")
	return nil
}

func RevokeOthers(c *gin.Context, userID, currentID uint64) error {
	now := time.Now()
	if err := authsessions.RevokeOtherTokens(userID, currentID, now); err != nil {
		return err
	}
	cachedSessions.InvalidateUser(userID)
	logEvent(c, authsessions.Token{UserId: userID, Id: currentID}, "revoke_others")
	return nil
}

func List(userID, version uint64) ([]authsessions.Token, error) {
	return authsessions.ListUserTokens(userID, version, time.Now())
}

func LogPasswordChange(c *gin.Context, userID uint64) {
	cachedSessions.InvalidateUser(userID)
	logEvent(c, authsessions.Token{UserId: userID}, "password_change")
}

func logEvent(c *gin.Context, session authsessions.Token, action string) {
	loginlogservice.Record(c, loginlogservice.Event{UserID: session.UserId, SessionID: session.Id,
		Action: action, Method: session.AuthMethod, Provider: session.OAuthProvider, Result: "success"})
}

func Cleanup() error {
	now := time.Now()
	if err := authsessions.CleanupTokens(now.Add(-24 * time.Hour)); err != nil {
		return err
	}
	return authsessions.CleanupLogs(now.Add(-90 * 24 * time.Hour))
}

const cleanupBatchSize = authsessions.CleanupBatchSize
