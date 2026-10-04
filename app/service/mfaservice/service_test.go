package mfaservice

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

func testUser(t *testing.T) users.EntityComplete {
	t.Helper()
	withSigningKey(t, testSigningKey(t))
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &usermfa.Factor{}, &usermfa.RecoveryCode{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	user := users.MakeUser(fmt.Sprintf("mfa-%d", time.Now().UnixNano()), "password123", "")
	user.IsActivated = users.ActivationSuccess
	if err := db.Create(user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.Id).Delete(&usermfa.Factor{})
		db.Where("user_id = ?", user.Id).Delete(&usermfa.RecoveryCode{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
		db.Unscoped().Delete(user)
	})
	if err := budgets.Clear(); err != nil {
		t.Fatal(err)
	}
	return *user
}
func testContext(cookie *http.Cookie) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/mfa/login", nil)
	if cookie != nil {
		c.Request.AddCookie(cookie)
	}
	return c, recorder
}
func enabledUser(t *testing.T) (users.EntityComplete, []string) {
	t.Helper()
	user := testUser(t)
	c, _ := testContext(nil)
	result, err := Begin(c, user.Id, "password123")
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	dbconnect.Connect().Model(&usermfa.Factor{}).Where("user_id = ?", user.Id).Count(&count)
	if count != 0 {
		t.Fatal("pending setup already enabled")
	}
	code, err := totp.GenerateCode(result["secret"].(string), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	codes, err := Change(c, user.Id, "password123", code, "enable")
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 10 {
		t.Fatal("wrong recovery count")
	}
	user.TokenVersion++
	return user, codes
}
func newChallenge(t *testing.T, user users.EntityComplete, method string) *http.Cookie {
	t.Helper()
	c, r := testContext(nil)
	required, err := CompleteFirstFactor(c, user.Id, user.TokenVersion, authsessionservice.LoginDetails{Method: method}, "/settings")
	if err != nil || !required {
		t.Fatalf("challenge: %v %v", required, err)
	}
	if r.Header().Get("New-Token") != "" {
		t.Fatal("session issued before MFA")
	}
	for _, cookie := range r.Result().Cookies() {
		if cookie.Name == challengeCookie {
			return cookie
		}
	}
	t.Fatal("missing challenge cookie")
	return nil
}
func TestStandardTOTPAndDrift(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	for _, second := range []int64{29, 59, 89} {
		step, err := acceptedStep(secret, "287082", time.Unix(second, 0))
		if err != nil || step != 1 {
			t.Fatalf("RFC code/drift at %d: %d %v", second, step, err)
		}
	}
	if _, err := acceptedStep(secret, "287082", time.Unix(119, 0)); err == nil {
		t.Fatal("accepted out-of-window code")
	}
}
func TestCipherBindingAndMissingKey(t *testing.T) {
	withSigningKey(t, testSigningKey(t))
	sealed, err := seal(1, "SECRET")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = open(2, sealed); !errors.Is(err, ErrUnavailable) {
		t.Fatal("cipher not bound to user")
	}
	preferences.Set("app.signingKey", "invalid")
	if _, err = open(1, sealed); !errors.Is(err, ErrUnavailable) {
		t.Fatal("missing key bypassed")
	}
}
func TestMFARecoveryChallengeAndReplay(t *testing.T) {
	user, codes := enabledUser(t)
	cookie := newChallenge(t, user, "oauth")
	c, r := testContext(cookie)
	redirect, err := Login(c, codes[0])
	if err != nil || redirect != "/settings" {
		t.Fatalf("login: %q %v", redirect, err)
	}
	if r.Header().Get("New-Token") == "" {
		t.Fatal("no session after MFA")
	}
	if _, err = Login(c, codes[0]); err == nil {
		t.Fatal("challenge replay accepted")
	}
	c, _ = testContext(newChallenge(t, user, "password"))
	if _, err = Login(c, codes[0]); err == nil {
		t.Fatal("recovery replay accepted")
	}
	status, err := Status(user.Id)
	if err != nil || status["remainingCodes"].(int64) != 9 {
		t.Fatalf("remaining: %v %v", status, err)
	}
}
func TestConcurrentRecoveryConsumption(t *testing.T) {
	user, codes := enabledUser(t)
	cookies := []*http.Cookie{newChallenge(t, user, "password"), newChallenge(t, user, "oauth")}
	var success atomic.Int32
	var wg sync.WaitGroup
	for _, cookie := range cookies {
		wg.Add(1)
		go func(cookie *http.Cookie) {
			defer wg.Done()
			c, _ := testContext(cookie)
			if _, err := Login(c, codes[0]); err == nil {
				success.Add(1)
			}
		}(cookie)
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatalf("same recovery code issued %d sessions", success.Load())
	}
}
func TestSessionFailureKeepsRecoveryConsumed(t *testing.T) {
	user, codes := enabledUser(t)
	cookie := newChallenge(t, user, "password")
	db := dbconnect.Connect()
	callback := "mfa_session_failure"
	if err := db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "user_auth_tokens" {
			tx.AddError(errors.New("forced session failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	c, r := testContext(cookie)
	_, err := Login(c, codes[0])
	db.Callback().Create().Remove(callback)
	if err == nil || r.Header().Get("New-Token") != "" {
		t.Fatal("failed session still issued")
	}
	var unused int64
	db.Model(&usermfa.RecoveryCode{}).Where("user_id = ? AND hash = ? AND used_at IS NULL", user.Id, codeHash(codes[0])).Count(&unused)
	if unused != 0 {
		t.Fatal("failed session restored recovery code")
	}
	if _, err = Login(c, codes[0]); err == nil {
		t.Fatal("consumed recovery code reused")
	}
	if _, err = Login(c, codes[1]); err != nil {
		t.Fatalf("retry with another recovery code: %v", err)
	}
}
func TestChangedPasswordInvalidatesChallenge(t *testing.T) {
	user, codes := enabledUser(t)
	cookie := newChallenge(t, user, "password")
	dbconnect.Connect().Model(&users.EntityComplete{}).Where("id = ?", user.Id).Update("token_version", user.TokenVersion+1)
	c, r := testContext(cookie)
	if _, err := Login(c, codes[0]); err == nil {
		t.Fatal("old challenge survived password change")
	}
	if r.Header().Get("New-Token") != "" {
		t.Fatal("issued session with old version")
	}
}
func TestTOTPStepConsumedOnce(t *testing.T) {
	user := testUser(t)
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	sealed, err := seal(user.Id, secret)
	if err != nil {
		t.Fatal(err)
	}
	if err = dbconnect.Connect().Create(&usermfa.Factor{UserID: user.Id, Secret: sealed, EnabledAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	code, _ := totp.GenerateCode(secret, time.Now())
	if err = consume(dbconnect.Connect(), user.Id, code); err != nil {
		t.Fatal(err)
	}
	if err = consume(dbconnect.Connect(), user.Id, code); err == nil {
		t.Fatal("TOTP replay accepted")
	}
}

func TestRegenerateThenDisableInvalidatesOldCodesAndChallenges(t *testing.T) {
	user, codes := enabledUser(t)
	stale := newChallenge(t, user, "oauth")
	c, _ := testContext(nil)
	replacement, err := Change(c, user.Id, "password123", codes[0], "regenerate")
	if err != nil || len(replacement) != 10 {
		t.Fatalf("regenerate: %d %v", len(replacement), err)
	}
	user.TokenVersion++
	c, r := testContext(stale)
	if _, err = Login(c, codes[1]); err == nil || r.Header().Get("New-Token") != "" {
		t.Fatal("old challenge survived regeneration")
	}
	if err = consume(dbconnect.Connect(), user.Id, codes[1]); err == nil {
		t.Fatal("old recovery code survived regeneration")
	}
	c, _ = testContext(nil)
	if _, err = Change(c, user.Id, "password123", replacement[0], "disable"); err != nil {
		t.Fatal(err)
	}
	user.TokenVersion++
	status, err := Status(user.Id)
	if err != nil || status["enabled"] != false || status["remainingCodes"] != int64(0) {
		t.Fatalf("disabled state: %v %v", status, err)
	}
	c, r = testContext(nil)
	required, err := CompleteFirstFactor(c, user.Id, user.TokenVersion, authsessionservice.LoginDetails{Method: "oauth"}, "/")
	if err != nil || required || r.Header().Get("New-Token") == "" {
		t.Fatalf("login after disable: required=%v err=%v", required, err)
	}
}

func TestExhaustedRecoveryCodesStillAllowTOTP(t *testing.T) {
	user, codes := enabledUser(t)
	for _, code := range codes {
		if err := consume(dbconnect.Connect(), user.Id, code); err != nil {
			t.Fatal(err)
		}
	}
	status, err := Status(user.Id)
	if err != nil || status["remainingCodes"] != int64(0) || status["enabled"] != true {
		t.Fatalf("exhausted state: %v %v", status, err)
	}
	factor, err := usermfa.GetFactor(user.Id)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := open(user.Id, factor.Secret)
	if err != nil {
		t.Fatal(err)
	}
	// Setup already consumed the current step; use the accepted future drift step.
	code, err := totp.GenerateCode(secret, time.Unix((factor.LastStep+1)*30, 0))
	if err != nil {
		t.Fatal(err)
	}
	c, r := testContext(newChallenge(t, user, "oauth"))
	if _, err = Login(c, code); err != nil || r.Header().Get("New-Token") == "" {
		t.Fatalf("TOTP after recovery exhaustion: %v", err)
	}
	if err = consume(dbconnect.Connect(), user.Id, codes[0]); err == nil {
		t.Fatal("exhausted recovery code accepted")
	}
}

func TestBannedAccountCannotFinishChallengeOrChangeMFA(t *testing.T) {
	user, codes := enabledUser(t)
	cookie := newChallenge(t, user, "oauth")
	if err := dbconnect.Connect().Model(&users.EntityComplete{}).Where("id = ?", user.Id).
		Update("restriction_status", users.RestrictionBanned).Error; err != nil {
		t.Fatal(err)
	}
	c, r := testContext(cookie)
	if _, err := Login(c, codes[0]); err == nil || r.Header().Get("New-Token") != "" {
		t.Fatal("banned account completed MFA")
	}
	if _, err := Change(c, user.Id, "password123", codes[0], "disable"); err == nil {
		t.Fatal("banned account disabled MFA")
	}
	if err := dbconnect.Connect().Model(&users.EntityComplete{}).Where("id = ?", user.Id).
		Update("restriction_status", users.RestrictionSuspended).Error; err != nil {
		t.Fatal(err)
	}
	c, _ = testContext(nil)
	if _, err := Change(c, user.Id, "password123", codes[0], "disable"); err != nil {
		t.Fatalf("suspended account cannot secure account: %v", err)
	}
}
