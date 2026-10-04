package authsessionservice

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func testContext() (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, recorder
}

func TestSessionLifecycle(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: "session-test-" + time.Now().Format("150405.000000000"), TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
		db.Unscoped().Delete(&user)
	})

	c, recorder := testContext()
	if err := Issue(c, user.Id, user.TokenVersion, LoginDetails{Method: "password", Reauthenticated: true}); err != nil {
		t.Fatal(err)
	}
	var raw string
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "access_token" {
			raw = cookie.Value
		}
	}
	if raw == "" {
		t.Fatal("no session cookie")
	}
	if got := recorder.Header().Get("New-Token"); got != raw {
		t.Fatalf("New-Token = %q, want issued token", got)
	}
	var session authsessions.Token
	if err := db.Where("user_id = ?", user.Id).First(&session).Error; err != nil {
		t.Fatal(err)
	}
	if session.TokenHash == raw || session.AuthId == "" || !session.Reauthenticated {
		t.Fatalf("session metadata invalid: %+v", session)
	}
	auth, err := Authenticate(c, raw, true)
	if err != nil || auth.Session.Id != session.Id {
		t.Fatalf("authenticate: %+v, %v", auth, err)
	}

	if err := db.Model(&session).Update("expires_at", time.Now().Add(30*time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	cachedSessions.InvalidateUser(user.Id)
	bearer, bearerRecorder := testContext()
	if _, err := Authenticate(bearer, raw, false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(bearerRecorder.Header().Get("Set-Cookie"), "access_token") {
		t.Fatal("bearer authentication wrote a cookie")
	}
	if err := db.First(&session, session.Id).Error; err != nil || session.ExpiresAt.Before(time.Now().Add(6*24*time.Hour)) {
		t.Fatalf("session not renewed: %+v, %v", session, err)
	}
	if err := db.Model(&session).Update("expires_at", time.Now().Add(30*time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	cachedSessions.InvalidateUser(user.Id)
	cookieContext, cookieRecorder := testContext()
	if _, err := Authenticate(cookieContext, raw, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cookieRecorder.Header().Get("Set-Cookie"), "access_token=") {
		t.Fatal("cookie renewal did not extend browser session")
	}
	forced := errors.New("revoke write failed")
	const callback = "test:auth_session_revoke_error"
	if err := db.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) { tx.AddError(forced) }); err != nil {
		t.Fatal(err)
	}
	if err := RevokeRaw(c, raw); !errors.Is(err, forced) {
		t.Fatalf("RevokeRaw error = %v", err)
	}
	if err := db.Callback().Update().Remove(callback); err != nil {
		t.Fatal(err)
	}
	if err := RevokeRaw(c, raw); err != nil {
		t.Fatal(err)
	}
	if _, err := Authenticate(c, raw, true); err == nil {
		t.Fatal("revoked session authenticated")
	}

	second, secondRecorder := testContext()
	if err := Issue(second, user.Id, user.TokenVersion, LoginDetails{Method: "oauth", Provider: "github"}); err != nil {
		t.Fatal(err)
	}
	var secondRaw string
	for _, cookie := range secondRecorder.Result().Cookies() {
		if cookie.Name == "access_token" {
			secondRaw = cookie.Value
		}
	}
	if secondRaw == "" {
		t.Fatal("second session cookie missing")
	}
	if _, err := Authenticate(second, secondRaw, true); err != nil {
		t.Fatal(err)
	}

	updated, err := users.UpdatePasswordByVersion(user.Id, 1, "changed")
	if err != nil || !updated {
		t.Fatalf("password update: %v, %v", updated, err)
	}
	if _, err := Authenticate(second, secondRaw, true); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("external password change must invalidate even a cached session: %v", err)
	}
	// In-process invalidation remains useful to release cached session data.
	LogPasswordChange(second, user.Id)
	if _, err := Authenticate(second, secondRaw, true); err == nil {
		t.Fatal("session survived password change")
	}
	if updated, err := users.UpdatePasswordByVersion(user.Id, 1, "again"); err != nil || updated {
		t.Fatalf("stale password update: %v, %v", updated, err)
	}
}

func TestCleanupDeletesExpiredRowsInBatches(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	userID := uint64(900000000 + now.UnixNano()%100000000)
	t.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&authsessions.Token{})
		db.Where("user_id = ?", userID).Delete(&authsessions.Log{})
	})
	tokens := make([]authsessions.Token, cleanupBatchSize+1)
	logs := make([]authsessions.Log, cleanupBatchSize+1)
	for i := range tokens {
		tokens[i] = authsessions.Token{UserId: userID, TokenHash: fmt.Sprintf("cleanup-%d-%d", now.UnixNano(), i), AuthMethod: "password", AuthTime: now, LastSeenAt: now, ExpiresAt: now.Add(-48 * time.Hour)}
		logs[i] = authsessions.Log{UserId: userID, Action: "generate", CreatedAt: now.Add(-91 * 24 * time.Hour)}
	}
	if err := db.CreateInBatches(tokens, 200).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.CreateInBatches(logs, 200).Error; err != nil {
		t.Fatal(err)
	}
	if err := Cleanup(); err != nil {
		t.Fatal(err)
	}
	var tokenCount, logCount int64
	db.Model(&authsessions.Token{}).Where("user_id = ?", userID).Count(&tokenCount)
	db.Model(&authsessions.Log{}).Where("user_id = ?", userID).Count(&logCount)
	if tokenCount != 0 || logCount != 0 {
		t.Fatalf("expired rows remain: tokens=%d logs=%d", tokenCount, logCount)
	}
}
