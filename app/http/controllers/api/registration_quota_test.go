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
	previousCaptcha := verifyRegistrationCaptcha
	verifyRegistrationCaptcha = func(_, code string) bool { return code == "valid" }
	t.Cleanup(func() { verifyRegistrationCaptcha = previousCaptcha })
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
