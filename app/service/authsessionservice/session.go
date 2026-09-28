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
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
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
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return err
	}
	raw := base64.RawURLEncoding.EncodeToString(secret)
	clear(secret)
	now := time.Now()
	session := authsessions.Token{
		UserId: userID, TokenHash: hash(raw), TokenVersion: tokenVersion,
		AuthMethod: details.Method, OAuthProvider: details.Provider,
		AuthTime: now, Reauthenticated: details.Reauthenticated,
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
	if err := dbconnect.Connect().Create(&session).Error; err != nil {
		return err
	}
	logEvent(c, session, "generate")
	// The limit is a convenience bound; validity always comes from the row lookup.
	var excess []authsessions.Token
	dbconnect.Connect().Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("last_seen_at DESC, id DESC").Offset(maxSessions).Find(&excess)
	for _, old := range excess {
		if revokeID(old.Id, userID, now) {
			cachedSessions.InvalidateUser(old.UserId)
			logEvent(c, old, "limit")
		}
	}
	SetCookie(c, raw, lifetime())
	return nil
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
	db := dbconnect.Connect()
	result := Authenticated{Session: session}
	needsRenewal := session.ExpiresAt.Sub(now) < renewalWindow()
	if needsRenewal {
		newExpiry := now.Add(lifetime())
		update := db.Model(&authsessions.Token{}).Where("id = ? AND revoked_at IS NULL AND expires_at > ? AND expires_at < ?", session.Id, now, newExpiry).
			Update("expires_at", newExpiry)
		if update.Error != nil {
			return Authenticated{}, update.Error
		}
		if update.RowsAffected > 0 {
			result.Renewed = true
			session.ExpiresAt = newExpiry
			cachedSessions.UpdateIfPresent(key, session)
		} else {
			cachedSessions.InvalidateUser(session.UserId)
		}
	}
	if session.LastSeenAt.Before(now.Add(-seenInterval)) {
		update := db.Model(&authsessions.Token{}).Where("id = ? AND revoked_at IS NULL AND last_seen_at < ?", session.Id, now.Add(-seenInterval)).Update("last_seen_at", now)
		if update.Error == nil && update.RowsAffected > 0 {
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
		var session authsessions.Token
		err := dbconnect.Connect().Model(&authsessions.Token{}).
			Joins("JOIN users ON users.id = user_auth_tokens.user_id AND users.deleted_at IS NULL AND users.token_version = user_auth_tokens.token_version").
			Where("user_auth_tokens.token_hash = ? AND user_auth_tokens.revoked_at IS NULL AND user_auth_tokens.expires_at > ?", key, now).
			First(&session).Error
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
	var session authsessions.Token
	err := dbconnect.Connect().Where("token_hash = ?", hash(raw)).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	result := dbconnect.Connect().Model(&authsessions.Token{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", session.Id, session.UserId).
		Update("revoked_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		cachedSessions.InvalidateUser(session.UserId)
		logEvent(c, session, "logout")
	}
	return nil
}

func revokeID(id, userID uint64, now time.Time) bool {
	result := dbconnect.Connect().Model(&authsessions.Token{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).Update("revoked_at", now)
	return result.Error == nil && result.RowsAffected == 1
}

func Revoke(c *gin.Context, userID, id, currentID uint64) error {
	if id == currentID {
		return ErrInvalidSession
	}
	var session authsessions.Token
	if err := dbconnect.Connect().Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).First(&session).Error; err != nil {
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
	result := dbconnect.Connect().Model(&authsessions.Token{}).Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, currentID).Update("revoked_at", now)
	if result.Error != nil {
		return result.Error
	}
	cachedSessions.InvalidateUser(userID)
	logEvent(c, authsessions.Token{UserId: userID, Id: currentID}, "revoke_others")
	return nil
}

func List(userID, version uint64) ([]authsessions.Token, error) {
	var sessions []authsessions.Token
	err := dbconnect.Connect().Where("user_id = ? AND token_version = ? AND revoked_at IS NULL AND expires_at > ?", userID, version, time.Now()).
		Order("last_seen_at DESC, id DESC").Find(&sessions).Error
	return sessions, err
}

func LogPasswordChange(c *gin.Context, userID uint64) {
	cachedSessions.InvalidateUser(userID)
	logEvent(c, authsessions.Token{UserId: userID}, "password_change")
}

func logEvent(c *gin.Context, session authsessions.Token, action string) {
	var ip, ua string
	if c != nil {
		ip, ua = c.ClientIP(), truncate(c.Request.UserAgent(), 512)
	}
	_ = dbconnect.Connect().Create(&authsessions.Log{UserId: session.UserId, UserAuthTokenId: session.Id, Action: action, ClientIP: ip, UserAgent: ua, CreatedAt: time.Now()}).Error
}

func Cleanup() error {
	db := dbconnect.Connect()
	now := time.Now()
	if err := deleteExpiredInBatches(db, &authsessions.Token{}, "expires_at < ? OR revoked_at < ?", now.Add(-24*time.Hour), now.Add(-24*time.Hour)); err != nil {
		return err
	}
	return deleteExpiredInBatches(db, &authsessions.Log{}, "created_at < ?", now.Add(-90*24*time.Hour))
}

const cleanupBatchSize = 500

func deleteExpiredInBatches(db *gorm.DB, model any, condition string, args ...any) error {
	for {
		var ids []uint64
		if err := db.Model(model).Where(condition, args...).Order("id").Limit(cleanupBatchSize).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if err := db.Where("id IN ?", ids).Delete(model).Error; err != nil {
			return err
		}
		if len(ids) < cleanupBatchSize {
			return nil
		}
	}
}
