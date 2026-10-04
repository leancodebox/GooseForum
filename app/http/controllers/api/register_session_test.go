package api

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func TestRegistrationIssuesSessionOnlyWithoutRequiredEmailVerification(t *testing.T) {
	useMailSettings(t, pageConfig.MailSettingsConfig{EnableMail: true, SmtpHost: "smtp.example.com", SmtpPort: 587, FromEmail: "forum@example.com"})
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}, &taskQueue.Entity{}, &userPoints.Entity{}, &pointsRecord.Entity{}, &userStatistics.Entity{}); err != nil {
		t.Fatal(err)
	}
	preserveSecuritySettings(t, db)
	originalCaptcha := verifyRegistrationCaptcha
	verifyRegistrationCaptcha = func(_, answer string) bool { return answer == "valid" }
	t.Cleanup(func() { verifyRegistrationCaptcha = originalCaptcha })
	// Registration is tested independently from first-user initialization.
	seed := users.EntityComplete{Id: uint64(800000000 + time.Now().UnixNano()%100000000), Username: fmt.Sprintf("register-seed-%d", time.Now().UnixNano())}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&seed) })

	for _, required := range []bool{true, false} {
		t.Run(fmt.Sprintf("email-verification-%t", required), func(t *testing.T) {
			config := fmt.Sprintf(`{"enableSignup":true,"enableEmailVerification":%t,"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`, required)
			if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, config); err != nil {
				t.Fatal(err)
			}
			hotdataserve.ClearSecuritySettingsConfigCache()
			suffix := time.Now().UnixNano()
			email := fmt.Sprintf("register-%d@example.com", suffix)
			payload, err := json.Marshal(vo.RegReq{Username: fmt.Sprintf("register-%d", suffix), Email: email, Password: "password123", CaptchaId: "id", CaptchaCode: "valid", Locale: "en"})
			if err != nil {
				t.Fatal(err)
			}
			c, response := mfaAPIContext(string(payload), nil)
			Register(c)
			var result component.ResultStruct
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			var user users.EntityComplete
			if err := db.Where("email = ?", email).First(&user).Error; err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
				db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
				db.Where("user_id = ?", user.Id).Delete(&userPoints.Entity{})
				db.Where("user_id = ?", user.Id).Delete(&pointsRecord.Entity{})
				db.Where("user_id = ?", user.Id).Delete(&userStatistics.Entity{})
				db.Where("task_json LIKE ?", "%"+email+"%").Delete(&taskQueue.Entity{})
				db.Unscoped().Delete(&user)
			})
			var sessions, mail int64
			if err := db.Model(&authsessions.Token{}).Where("user_id = ?", user.Id).Count(&sessions).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Model(&taskQueue.Entity{}).Where("task_json LIKE ?", "%"+email+"%").Count(&mail).Error; err != nil {
				t.Fatal(err)
			}
			if result.Code != component.SUCCESS {
				t.Fatalf("registration failed: %+v", result)
			}
			if required {
				if result.MessageCode != component.MessageAuthRegisterEmailVerify || sessions != 0 || response.Header().Get("New-Token") != "" || len(response.Result().Cookies()) != 0 {
					t.Fatalf("pending registration signed in: response=%+v sessions=%d headers=%v", result, sessions, response.Header())
				}
				if !user.RequiresEmailVerification || user.IsActivated != users.ActivationPending || mail != 1 {
					t.Fatalf("missing verification state: pending=%t activated=%d queued=%d", user.RequiresEmailVerification, user.IsActivated, mail)
				}
			} else if result.MessageCode != component.MessageAuthLoginSuccess || sessions != 1 || response.Header().Get("New-Token") == "" || len(response.Result().Cookies()) == 0 || user.IsActivated != users.ActivationSuccess {
				t.Fatalf("normal registration did not sign in: response=%+v sessions=%d activated=%d", result, sessions, user.IsActivated)
			}
		})
	}
}
