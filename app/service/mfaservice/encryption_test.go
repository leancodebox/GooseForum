package mfaservice

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/viper"
)

func testSigningKey(t *testing.T) string {
	t.Helper()
	key, err := algorithm.GenerateSigningKey(32)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func withSigningKey(t *testing.T, key string) {
	t.Helper()
	old := preferences.GetString("app.signingKey", "")
	preferences.Set("app.signingKey", key)
	t.Cleanup(func() { preferences.Set("app.signingKey", old) })
}

func TestGeneratedSigningKeySupportsEncryption(t *testing.T) {
	config, err := preferences.GenerateConfig()
	if err != nil {
		t.Fatal(err)
	}
	parsed := viper.New()
	parsed.SetConfigType("toml")
	if err := parsed.ReadConfig(bytes.NewReader(config)); err != nil {
		t.Fatal(err)
	}
	key := parsed.GetString("app.signingKey")
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(key)
	if err != nil || len(key) != 43 || len(decoded) != 32 {
		t.Fatal("generated configuration must contain a canonical 256-bit signing key")
	}
	if parsed.IsSet("mfa") {
		t.Fatal("configuration must only use the application signing key")
	}
	withSigningKey(t, key)
	sealed, err := seal(1, "SECRET")
	if err != nil || !strings.HasPrefix(sealed, "v1:") {
		t.Fatalf("derived encryption failed: %v", err)
	}
	if plain, err := open(1, sealed); err != nil || plain != "SECRET" {
		t.Fatalf("derived decryption failed: %v", err)
	}
	for _, input := range []string{strings.Replace(sealed, "v1:", "v2:", 1), "v1:invalid", "invalid"} {
		if _, err := open(1, input); !errors.Is(err, ErrUnavailable) {
			t.Fatal("accepted malformed ciphertext")
		}
	}
	if _, err := open(2, sealed); !errors.Is(err, ErrUnavailable) {
		t.Fatal("ciphertext was not bound to its user")
	}
	preferences.Set("app.signingKey", testSigningKey(t))
	if _, err := open(1, sealed); !errors.Is(err, ErrUnavailable) {
		t.Fatal("changing signing key did not invalidate ciphertext")
	}
}

func TestEncryptionRejectsInvalidSigningKeys(t *testing.T) {
	withSigningKey(t, testSigningKey(t))
	for _, value := range []string{"", "short", strings.Repeat("!", 43), strings.Repeat("a", 43),
		base64.RawURLEncoding.EncodeToString(make([]byte, 31)), base64.RawURLEncoding.EncodeToString(make([]byte, 33))} {
		preferences.Set("app.signingKey", value)
		if _, err := seal(1, "SECRET"); !errors.Is(err, ErrUnavailable) {
			t.Fatal("invalid signing key accepted")
		}
	}
}

func TestStatusChecksTheKeyOfAnExistingFactor(t *testing.T) {
	user, _ := enabledUser(t)
	preferences.Set("app.signingKey", testSigningKey(t))
	status, err := Status(user.Id)
	if err != nil || status["enabled"] != true || status["available"] != false {
		t.Fatalf("existing factor falsely available after key change: %v %v", status, err)
	}
}

func TestMFASetupAndLoginUsingOnlySigningKey(t *testing.T) {
	user := testUser(t)
	c, _ := testContext(nil)
	setup, err := Begin(c, user.Id, "password123")
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(setup["secret"].(string), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	codes, err := Change(c, user.Id, "password123", code, "enable")
	if err != nil || len(codes) != 10 {
		t.Fatalf("setup failed: %v", err)
	}
	status, err := Status(user.Id)
	if err != nil || status["enabled"] != true || status["available"] != true {
		t.Fatalf("status: %v %v", status, err)
	}
	user.TokenVersion++
	c, recorder := testContext(newChallenge(t, user, "password"))
	if _, err := Login(c, codes[0]); err != nil || recorder.Header().Get("New-Token") == "" {
		t.Fatalf("login failed: %v", err)
	}
}
