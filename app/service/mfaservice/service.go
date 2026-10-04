package mfaservice

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/leancodebox/GooseForum/app/service/loginlogservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

var ErrVerification = errors.New("verification failed or expired")
var ErrUnavailable = errors.New("MFA encryption key is unavailable; contact the administrator")

const challengeCookie = "mfa_challenge"

type challenge struct {
	UserID, Version  uint64
	IP, UA, Redirect string
	Details          authsessionservice.LoginDetails
	Expires          time.Time
	Attempts         int
	Busy             bool
}
type pending struct {
	Secret  string
	Version uint64
	Expires time.Time
}
type budget struct {
	Attempts int
	Expires  time.Time
}

var challenges = &sharedcache.Cache[challenge]{Name: "mfa.challenges", MaxEntries: 10000}
var bindings = &sharedcache.Cache[pending]{Name: "mfa.pending", MaxEntries: 10000}
var budgets = &sharedcache.Cache[budget]{Name: "mfa.budgets", MaxEntries: 10000}

func userStateKey(userID uint64) string { return fmt.Sprintf("%d:", userID) }

func clearUserState(userID uint64) error {
	if err := bindings.Delete(userStateKey(userID)); err != nil {
		return err
	}
	return challenges.DeletePrefix(userStateKey(userID))
}

func encryption() (cipher.AEAD, error) {
	value := preferences.GetString("app.signingKey", "")
	if len(value) != 43 {
		return nil, ErrUnavailable
	}
	secret, err := base64.RawURLEncoding.Strict().DecodeString(value)
	defer clear(secret)
	if err != nil || len(secret) != 32 {
		return nil, ErrUnavailable
	}
	key, err := hkdf.Key(sha256.New, secret, nil, "gooseforum/mfa/totp/v1", 32)
	defer clear(key)
	if err != nil || len(key) != 32 {
		return nil, ErrUnavailable
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrUnavailable
	}
	return cipher.NewGCM(block)
}
func seal(userID uint64, secret string) (string, error) {
	a, err := encryption()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, a.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	prefix := "v1:"
	ciphertext := a.Seal(nonce, nonce, []byte(secret), []byte(fmt.Sprintf("mfa:%s%d", prefix, userID)))
	return prefix + base64.RawStdEncoding.EncodeToString(ciphertext), nil
}
func open(userID uint64, secret string) (string, error) {
	if !strings.HasPrefix(secret, "v1:") {
		return "", ErrUnavailable
	}
	a, err := encryption()
	if err != nil {
		return "", err
	}
	data, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(secret, "v1:"))
	if err != nil || len(data) < a.NonceSize() {
		return "", ErrUnavailable
	}
	plain, err := a.Open(nil, data[:a.NonceSize()], data[a.NonceSize():], []byte(fmt.Sprintf("mfa:v1:%d", userID)))
	if err != nil {
		return "", ErrUnavailable
	}
	return string(plain), nil
}
func random() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	defer clear(b)
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func allow(userID uint64, ip string) bool {
	now := time.Now()
	keys := []string{fmt.Sprintf("user:%d", userID), "ip:" + ip}
	for index, k := range keys {
		limit := 10
		if index == 1 {
			limit = 100
		}
		err := budgets.AtomicUpdate(k, func(v budget, found bool) (budget, time.Duration, error) {
			if !found || !v.Expires.After(now) {
				v = budget{Expires: now.Add(5 * time.Minute)}
			}
			if v.Attempts >= limit {
				return v, 0, ErrVerification
			}
			v.Attempts++
			return v, time.Until(v.Expires), nil
		})
		if err != nil {
			return false
		}
	}
	return true
}
func policy(db *gorm.DB, userID, version uint64) (users.MFAAuthState, error) {
	user, err := users.GetMFAAuthStateWithDB(db, userID, version)
	if err != nil {
		return user, ErrVerification
	}
	if user.EffectiveRestriction(time.Now()) == users.RestrictionBanned {
		return user, ErrVerification
	}
	if user.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) {
		return user, ErrVerification
	}
	return user, nil
}
func CompleteFirstFactor(c *gin.Context, userID, version uint64, details authsessionservice.LoginDetails, redirect string) (bool, error) {
	details.AuthTime = time.Now()
	if _, err := policy(dbconnect.Connect(), userID, version); err != nil {
		return false, err
	}
	factor, err := usermfa.GetFactor(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, authsessionservice.Issue(c, userID, version, details)
	}
	if err != nil {
		return false, err
	}
	if _, err := open(userID, factor.Secret); err != nil {
		return false, err
	}
	raw, err := random()
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") || strings.ContainsAny(redirect, "\\\r\n") {
		redirect = "/"
	}
	raw = userStateKey(userID) + raw
	if err := challenges.Set(raw, challenge{UserID: userID, Version: version, IP: c.ClientIP(), UA: c.Request.UserAgent(), Details: details, Redirect: redirect, Expires: time.Now().Add(5 * time.Minute)}, 5*time.Minute); err != nil {
		return false, err
	}
	authsessionservice.ClearCookie(c)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(challengeCookie, raw, 300, "/api/mfa/login", "", !setting.IsLocal(), true)
	loginlogservice.Record(c, loginlogservice.Event{Action: "mfa_challenge", Method: details.Method, Provider: details.Provider, Result: "challenge"})
	return true, nil
}

func Login(c *gin.Context, code string) (string, error) {
	raw, _ := c.Cookie(challengeCookie)
	var item challenge
	err := challenges.AtomicUpdate(raw, func(current challenge, found bool) (challenge, time.Duration, error) {
		if !found || !current.Expires.After(time.Now()) || current.IP != c.ClientIP() || current.UA != c.Request.UserAgent() || current.Busy || current.Attempts >= 5 {
			return current, 0, ErrVerification
		}
		current.Attempts++
		current.Busy = true
		item = current
		return current, time.Until(current.Expires), nil
	})
	if err != nil {
		loginlogservice.Failure(c, "mfa", "", "challenge_invalid")
		return "", ErrVerification
	}
	if !allow(item.UserID, c.ClientIP()) {
		_ = challenges.Delete(raw)
		return "", ErrVerification
	}
	err = authsessionservice.IssueVerified(c, item.UserID, item.Version, item.Details, func(tx *gorm.DB) error {
		if _, err := policy(tx, item.UserID, item.Version); err != nil {
			return err
		}
		return consume(tx, item.UserID, code)
	})
	if err != nil {
		_ = challenges.AtomicUpdate(raw, func(current challenge, found bool) (challenge, time.Duration, error) {
			if !found {
				return current, 0, nil
			}
			current.Busy = false
			return current, time.Until(current.Expires), nil
		})
		loginlogservice.Record(c, loginlogservice.Event{Action: "challenge_failure", Method: "mfa", Provider: item.Details.Provider, Result: "failure", Reason: "code_rejected"})
		return "", err
	}
	_ = challenges.Delete(raw)
	c.SetCookie(challengeCookie, "", -1, "/api/mfa/login", "", !setting.IsLocal(), true)
	return item.Redirect, nil
}
func acceptedStep(secret, code string, now time.Time) (int64, error) {
	step := now.Unix() / 30
	for _, offset := range []int64{0, -1, 1} {
		valid, err := totp.ValidateCustom(code, secret, time.Unix((step+offset)*30, 0), totp.ValidateOpts{Period: 30, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
		if err == nil && valid {
			return step + offset, nil
		}
	}
	return 0, ErrVerification
}
func codeHash(code string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
func consume(tx *gorm.DB, userID uint64, code string) error {
	factor, err := usermfa.GetFactorWithDB(tx, userID)
	if err != nil {
		return ErrVerification
	}
	if len(code) == 6 {
		secret, err := open(userID, factor.Secret)
		if err != nil {
			return err
		}
		step, err := acceptedStep(secret, code, time.Now())
		if err != nil {
			return err
		}
		accepted, err := usermfa.AcceptStepWithDB(tx, userID, step)
		if err != nil {
			return err
		}
		if !accepted {
			return ErrVerification
		}
		return nil
	}
	accepted, err := usermfa.ConsumeRecoveryCodeWithDB(tx, userID, codeHash(code), time.Now())
	if err != nil {
		return err
	}
	if !accepted {
		return ErrVerification
	}
	return nil
}
func Status(userID uint64) (map[string]any, error) {
	factor, err := usermfa.GetFactor(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	enabled := err == nil
	remaining, err := usermfa.CountUnusedRecoveryCodes(userID)
	if err != nil {
		return nil, err
	}
	_, keyErr := encryption()
	if enabled {
		_, keyErr = open(userID, factor.Secret)
	}
	return map[string]any{"enabled": enabled, "available": keyErr == nil, "remainingCodes": remaining}, nil
}
func verifyPassword(userID uint64, password string) (users.EntityComplete, error) {
	user, err := users.Get(userID)
	if err != nil {
		return user, ErrVerification
	}
	if err = algorithm.VerifyEncryptPassword(user.Password, password); err != nil {
		return user, ErrVerification
	}
	_, err = policy(dbconnect.Connect(), userID, user.TokenVersion)
	return user, err
}

// VerifySecondFactor protects account security changes when MFA is enabled.
func VerifySecondFactor(c *gin.Context, userID uint64, code string) (err error) {
	defer func() {
		if err != nil {
			loginlogservice.Record(c, loginlogservice.Event{UserID: userID, Action: "security_failure", Method: "mfa", Result: "failure", Reason: "second_factor_rejected"})
		}
	}()
	enabled, err := usermfa.HasFactor(userID)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	allowed := allow(userID, c.ClientIP())
	if !allowed || len(code) > 128 {
		return ErrVerification
	}
	return consume(dbconnect.Connect(), userID, strings.TrimSpace(code))
}
func Begin(c *gin.Context, userID uint64, password string) (map[string]any, error) {
	allowed := allow(userID, c.ClientIP())
	if !allowed {
		return nil, ErrVerification
	}
	user, err := verifyPassword(userID, password)
	if err != nil {
		return nil, err
	}
	status, err := Status(userID)
	if err != nil {
		return nil, err
	}
	if status["enabled"] == true {
		return nil, ErrVerification
	}
	if _, err = encryption(); err != nil {
		return nil, err
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "GooseForum", AccountName: user.Username, Period: 30, Digits: otp.DigitsSix, SecretSize: 20})
	if err != nil {
		return nil, err
	}
	secret, err := seal(userID, key.Secret())
	if err != nil {
		return nil, err
	}
	if err := bindings.Set(userStateKey(userID), pending{Secret: secret, Version: user.TokenVersion, Expires: time.Now().Add(10 * time.Minute)}, 10*time.Minute); err != nil {
		return nil, err
	}
	return map[string]any{"secret": key.Secret(), "uri": key.URL()}, nil
}
func recoveryCodes(tx *gorm.DB, userID uint64) ([]string, error) {
	if err := usermfa.DeleteRecoveryCodesWithDB(tx, userID); err != nil {
		return nil, err
	}
	codes := make([]string, 10)
	rows := make([]usermfa.RecoveryCode, 10)
	for i := range codes {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		h := strings.ToUpper(hex.EncodeToString(b))
		codes[i] = h[:8] + "-" + h[8:16] + "-" + h[16:24] + "-" + h[24:]
		rows[i] = usermfa.RecoveryCode{UserID: userID, Hash: codeHash(codes[i])}
		clear(b)
	}
	return codes, usermfa.CreateRecoveryCodesWithDB(tx, rows)
}
func Change(c *gin.Context, userID uint64, password, code, action string) ([]string, error) {
	allowed := allow(userID, c.ClientIP())
	item, ok, cacheErr := bindings.Get(userStateKey(userID))
	if cacheErr != nil {
		return nil, cacheErr
	}
	if !allowed {
		return nil, ErrVerification
	}
	user, err := verifyPassword(userID, password)
	if err != nil {
		return nil, err
	}
	var codes []string
	err = dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		if _, err := policy(tx, userID, user.TokenVersion); err != nil {
			return err
		}
		if action == "enable" {
			if !ok || item.Version != user.TokenVersion || !item.Expires.After(time.Now()) {
				return ErrVerification
			}
			secret, err := open(userID, item.Secret)
			if err != nil {
				return err
			}
			step, err := acceptedStep(secret, code, time.Now())
			if err != nil {
				return err
			}
			if err = usermfa.CreateFactorWithDB(tx, &usermfa.Factor{UserID: userID, Secret: item.Secret, EnabledAt: time.Now(), LastStep: step}); err != nil {
				return err
			}
		} else if action == "disable" || action == "regenerate" {
			if err := consume(tx, userID, code); err != nil {
				return err
			}
		} else {
			return ErrVerification
		}
		if action == "disable" {
			if _, err := usermfa.DeleteFactorWithDB(tx, userID); err != nil {
				return err
			}
			if err := usermfa.DeleteRecoveryCodesWithDB(tx, userID); err != nil {
				return err
			}
		} else {
			var err error
			codes, err = recoveryCodes(tx, userID)
			if err != nil {
				return err
			}
		}
		advanced, err := users.CompareAndAdvanceTokenVersionWithDB(tx, userID, user.TokenVersion)
		if err != nil {
			return err
		}
		if !advanced {
			return ErrVerification
		}
		return authsessions.RevokeAllTokensWithDB(tx, userID, time.Now())
	})
	if err != nil {
		return nil, err
	}
	_ = clearUserState(userID)
	authsessionservice.InvalidateUser(userID)
	if current, err := users.Get(userID); err == nil {
		userservice.RefreshUserCaches(&current)
	}
	authsessionservice.ClearCookie(c)
	authsessionservice.LogSecurityEvent(c, userID, "mfa_"+action)
	return codes, nil
}
