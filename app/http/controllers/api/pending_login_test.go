package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/logincrypto"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/kvstore"
)

func loginPasswordCipher(t *testing.T, password string) string {
	t.Helper()
	block, _ := pem.Decode([]byte(logincrypto.PublicKeyPEM()))
	if block == nil {
		t.Fatal("missing login public key")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(logincrypto.PasswordPayload{Password: password, Ts: time.Now().UnixMilli()})
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key.(*rsa.PublicKey), payload, nil)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func TestPendingPasswordLoginResendsWithoutIssuingSession(t *testing.T) {
	useMailSettings(t, pageConfig.MailSettingsConfig{EnableMail: true, SmtpHost: "smtp.example.com", SmtpPort: 587, FromEmail: "forum@example.com"})
	originalCaptcha := verifyLoginCaptcha
	verifyLoginCaptcha = func(_, answer string) bool { return answer == "valid" }
	t.Cleanup(func() { verifyLoginCaptcha = originalCaptcha })
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &taskQueue.Entity{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	user := users.MakeUser(fmt.Sprintf("pending-%d", time.Now().UnixNano()), "password123", fmt.Sprintf("pending-%d@example.com", time.Now().UnixNano()))
	user.RequiresEmailVerification = true
	if err := db.Create(user).Error; err != nil {
		t.Fatal(err)
	}
	quotaKey := "activation-email:resend:" + strconv.FormatUint(user.Id, 10) + ":" + time.Now().Format("20060102")
	if err := kvstore.Set(quotaKey, `{"count":0}`, time.Minute); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = kvstore.Set(quotaKey, "", time.Nanosecond) })
	t.Cleanup(func() {
		db.Where("task_json LIKE ?", "%"+user.Email+"%").Delete(&taskQueue.Entity{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
		db.Unscoped().Delete(user)
	})
	login := func(password, captcha string) component.ResultStruct {
		t.Helper()
		payload, err := json.Marshal(LoginReq{Username: user.Username, EncryptedPassword: loginPasswordCipher(t, password), CaptchaId: "id", CaptchaCode: captcha})
		if err != nil {
			t.Fatal(err)
		}
		c, response := mfaAPIContext(string(payload), nil)
		Login(c)
		if response.Header().Get("New-Token") != "" {
			t.Fatal("pending login issued session")
		}
		var result component.ResultStruct
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	if result := login("password123", "invalid"); result.MessageCode != component.MessageAuthCaptchaInvalid {
		t.Fatalf("captcha: %+v", result)
	}
	if result := login("wrong-password", "valid"); result.MessageCode != component.MessageAuthInvalidCredentials {
		t.Fatalf("password: %+v", result)
	}
	var queued int64
	db.Model(&taskQueue.Entity{}).Where("task_json LIKE ?", "%"+user.Email+"%").Count(&queued)
	if queued != 0 {
		t.Fatal("unverified first factor sent mail")
	}
	if result := login("password123", "valid"); result.Code != component.FAIL || result.MessageCode != component.MessageAuthActivationResendSuccess {
		t.Fatalf("verified pending login: %+v", result)
	}
	if result := login("password123", "valid"); result.MessageCode != component.MessageAuthActivationResendCooldown {
		t.Fatalf("cooldown: %+v", result)
	}
	db.Model(&taskQueue.Entity{}).Where("task_json LIKE ?", "%"+user.Email+"%").Count(&queued)
	if queued != 1 {
		t.Fatalf("queued %d emails, want 1", queued)
	}
	var sessions int64
	db.Model(&authsessions.Token{}).Where("user_id = ?", user.Id).Count(&sessions)
	if sessions != 0 {
		t.Fatal("unverified login created a session row")
	}
}

func TestPendingLoginMissingMailReportsFailure(t *testing.T) {
	useMailSettings(t, pageConfig.MailSettingsConfig{})
	user := users.EntityComplete{Id: 98712939, Email: fmt.Sprintf("pending-missing-smtp-%d@example.com", time.Now().UnixNano()), RequiresEmailVerification: true}
	c, response := mfaAPIContext("", nil)
	result := resendPendingLogin(c, user)
	if result.Code != component.FAIL || result.MessageCode != component.MessageAuthActivationResendFailed {
		t.Fatalf("missing mail: %+v", result)
	}
	if response.Header().Get("New-Token") != "" {
		t.Fatal("missing SMTP bypassed verification")
	}
}

func TestPasswordRegistrationMissingMailDoesNotCreateAccount(t *testing.T) {
	useMailSettings(t, pageConfig.MailSettingsConfig{})
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	previous := pageConfig.GetByPageType(pageConfig.SecuritySettings)
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, `{"enableSignup":true,"enableEmailVerification":true,"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
	t.Cleanup(func() {
		if previous.Id != 0 {
			if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, previous.Config); err != nil {
				t.Error(err)
			}
		} else if err := db.Where("page_type = ?", pageConfig.SecuritySettings).Delete(&pageConfig.Entity{}).Error; err != nil {
			t.Error(err)
		}
		hotdataserve.ClearSecuritySettingsConfigCache()
	})
	email := fmt.Sprintf("register-missing-mail-%d@example.com", time.Now().UnixNano())
	body := fmt.Sprintf(`{"userName":"mail-check","email":%q,"passWord":"password123","captchaId":"id","captchaCode":"code"}`, email)
	c, response := mfaAPIContext(body, nil)
	Register(c)
	var result component.ResultStruct
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Code != component.FAIL || result.MessageCode != component.MessageAuthActivationResendFailed {
		t.Fatalf("missing mail registration: %+v", result)
	}
	if users.ExistEmail(email) || response.Header().Get("New-Token") != "" {
		t.Fatal("missing mail created an account or session")
	}
}
