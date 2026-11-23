package api

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func TestInvalidRegistrationCaptchaCannotExhaustSignupQuota(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	preserveSecuritySettings(t, db)
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, `{"enableSignup":true,"enableEmailVerification":false,"registrationIPLimit":0,"registrationEmailLimit":1,"registrationGlobalLimit":0}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
	previousCaptcha := verifyImageCaptcha
	verifyImageCaptcha = func(_, code string) bool { return code == "valid" }
	t.Cleanup(func() { verifyImageCaptcha = previousCaptcha })
	stamp := time.Now().UnixNano()
	username := fmt.Sprintf("quota%d", stamp)
	existing := users.EntityComplete{Username: username, Email: fmt.Sprintf("quota-existing-%d@example.com", stamp)}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&existing) })
	email := fmt.Sprintf("quota-target-%d@example.com", stamp)
	request := func(captcha string) component.ResultStruct {
		t.Helper()
		body := fmt.Sprintf(`{"userName":%q,"email":%q,"passWord":"password123","captchaId":"id","captchaCode":%q}`, username, email, captcha)
		context, response := mfaAPIContext(body, nil)
		Register(context)
		var result component.ResultStruct
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	for range 3 {
		if result := request("invalid"); result.MessageCode != component.MessageAuthCaptchaInvalid {
			t.Fatalf("invalid captcha consumed signup quota: %+v", result)
		}
	}
	if result := request("valid"); result.MessageCode != component.MessageAuthUsernameExists {
		t.Fatalf("authenticated signup attempt did not retain its quota: %+v", result)
	}
	if result := request("valid"); result.MessageCode != component.MessageRegistrationRateLimited {
		t.Fatalf("valid attempts bypassed signup quota: %+v", result)
	}
}

func TestRegistrationUsesServerCaptchaPolicy(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	preserveSecuritySettings(t, db)
	stamp := time.Now().UnixNano()
	user := users.EntityComplete{Username: fmt.Sprintf("policy%d", stamp), Email: fmt.Sprintf("policy%d@example.com", stamp)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&user) })
	for _, tc := range []struct {
		provider string
		want     component.MessageCode
	}{
		{"image", component.MessageAuthCaptchaInvalid},
		{"hcaptcha", component.MessageAuthCaptchaInvalid},
		{"recaptcha_v2", component.MessageAuthCaptchaInvalid},
		{"recaptcha_v3", component.MessageAuthCaptchaInvalid},
		{"none", component.MessageAuthUsernameExists},
	} {
		t.Run(tc.provider, func(t *testing.T) {
			settings := fmt.Sprintf(`{"enableSignup":true,"registrationCaptcha":{"provider":%q},"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`, tc.provider)
			if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, settings); err != nil {
				t.Fatal(err)
			}
			hotdataserve.ClearSecuritySettingsConfigCache()
			req, rec := mfaAPIContext(fmt.Sprintf(`{"userName":%q,"email":%q,"passWord":"password123","provider":"none"}`, user.Username, user.Email), nil)
			Register(req)
			var result component.ResultStruct
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.MessageCode != tc.want {
				t.Errorf("got %s, want %s", result.MessageCode, tc.want)
			}
		})
	}
}

func TestPasswordLoginCaptchaAndAttemptLimit(t *testing.T) {
	db := dbconnect.Connect()
	preserveSecuritySettings(t, db)
	cipher := loginPasswordCipher(t, "password123")
	attempt := func(peer, encrypted string) component.ResultStruct {
		t.Helper()
		c, w := mfaAPIContext(fmt.Sprintf(`{"username":"login-cap-test","encryptedPassword":%q}`, encrypted), nil)
		c.Request.RemoteAddr = peer
		Login(c)
		var result component.ResultStruct
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	stamp := uint64(time.Now().UnixNano())
	policyPeer := fmt.Sprintf("[2001:db8:%x:%x::1]:1234", uint16(stamp>>16), uint16(stamp))
	quotaPeer := fmt.Sprintf("[2001:db8:%x:%x::2]:1234", uint16(stamp>>16), uint16(stamp))
	for _, provider := range []string{"image", "hcaptcha", "recaptcha_v2", "recaptcha_v3", "none"} {
		if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, fmt.Sprintf(`{"registrationCaptcha":{"provider":%q}}`, provider)); err != nil {
			t.Fatal(err)
		}
		hotdataserve.ClearSecuritySettingsConfigCache()
		want := component.MessageAuthCaptchaInvalid
		if provider == "none" {
			want = component.MessageAuthInvalidCredentials
		}
		if got := attempt(policyPeer, cipher); got.MessageCode != want {
			t.Fatalf("%s: got %s, want %s", provider, got.MessageCode, want)
		}
	}
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, `{"registrationCaptcha":{"provider":"none"},"enableSignup":"malformed"}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
	if got := attempt(policyPeer, cipher); got.MessageCode != component.MessageAuthLoginFailed {
		t.Fatalf("login accepted invalid policy: %s", got.MessageCode)
	}

	for range 10 {
		if got := attempt(quotaPeer, "invalid"); got.MessageCode != component.MessageAuthLoginInvalidRequest {
			t.Fatalf("unexpected %s", got.MessageCode)
		}
	}
	if got := attempt(quotaPeer, "invalid"); got.MessageCode != component.MessageLoginRateLimited {
		t.Fatalf("expected limit before decryption, got %s", got.MessageCode)
	}
}
