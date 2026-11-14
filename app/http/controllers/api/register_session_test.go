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
	"github.com/leancodebox/GooseForum/app/models/forum/invitations"
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

func TestRegistrationInvitationCommitsExactlyOnce(t *testing.T) {
	useMailSettings(t, pageConfig.MailSettingsConfig{EnableMail: true, SmtpHost: "smtp.example.com", SmtpPort: 587, FromEmail: "forum@example.com"})
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &invitations.Entity{}, &userPoints.Entity{}, &pointsRecord.Entity{}, &userStatistics.Entity{}); err != nil {
		t.Fatal(err)
	}
	preserveSecuritySettings(t, db)
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, `{"enableSignup":true,"enableEmailVerification":true,"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
	previousCaptcha := verifyRegistrationCaptcha
	verifyRegistrationCaptcha = func(_, answer string) bool { return answer == "valid" }
	t.Cleanup(func() { verifyRegistrationCaptcha = previousCaptcha })
	seed := users.EntityComplete{Id: uint64(810000000 + time.Now().UnixNano()%100000000), Username: fmt.Sprintf("invite-seed-%d", time.Now().UnixNano())}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&seed) })
	_, code, err := invitations.CreateWithDB(db, seed.Id, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UnixNano()
	email := fmt.Sprintf("invited-%d@example.com", stamp)
	payload, err := json.Marshal(vo.RegReq{Username: fmt.Sprintf("invited-%d", stamp), Email: email, Password: "password123", InvitationCode: code, CaptchaId: "id", CaptchaCode: "valid"})
	if err != nil {
		t.Fatal(err)
	}
	c, response := mfaAPIContext(string(payload), nil)
	Register(c)
	var result component.ResultStruct
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Code != component.SUCCESS {
		t.Fatalf("invited registration failed: %+v", result)
	}
	var invitation invitations.Entity
	if err := db.Where("hash = ?", invitations.Digest(code)).First(&invitation).Error; err != nil {
		t.Fatal(err)
	}
	if invitation.UsedCount != 1 || invitation.ReservationHash != "" || invitation.LastUsedUserId == 0 {
		t.Fatalf("invitation after registration = %+v", invitation)
	}
	var user users.EntityComplete
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.Id).Delete(&userPoints.Entity{})
		db.Where("user_id = ?", user.Id).Delete(&pointsRecord.Entity{})
		db.Where("user_id = ?", user.Id).Delete(&userStatistics.Entity{})
		db.Unscoped().Delete(&user)
	})
	secondPayload, _ := json.Marshal(vo.RegReq{Username: fmt.Sprintf("invitedsecond%d", stamp), Email: fmt.Sprintf("invited-second-%d@example.com", stamp), Password: "password123", InvitationCode: code, CaptchaId: "id", CaptchaCode: "valid"})
	secondContext, secondResponse := mfaAPIContext(string(secondPayload), nil)
	Register(secondContext)
	var secondResult component.ResultStruct
	if err := json.Unmarshal(secondResponse.Body.Bytes(), &secondResult); err != nil {
		t.Fatal(err)
	}
	if secondResult.MessageCode != component.MessageAuthInvitationInvalid {
		t.Fatalf("second invited registration = %+v, want invalid invitation", secondResult)
	}
}

func TestInviteOnlyRegistrationRequiresCode(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	preserveSecuritySettings(t, db)
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, `{"enableSignup":true,"inviteOnly":true,"enableEmailVerification":false,"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
	previousCaptcha := verifyRegistrationCaptcha
	verifyRegistrationCaptcha = func(_, answer string) bool { return answer == "valid" }
	t.Cleanup(func() { verifyRegistrationCaptcha = previousCaptcha })
	stamp := time.Now().UnixNano()
	payload, _ := json.Marshal(vo.RegReq{Username: "requiresinvite", Email: fmt.Sprintf("requiresinvite%d@example.com", stamp), Password: "password123", CaptchaId: "id", CaptchaCode: "valid"})
	c, response := mfaAPIContext(string(payload), nil)
	Register(c)
	var result component.ResultStruct
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.MessageCode != component.MessageAuthInvitationRequired {
		t.Fatalf("invite-only registration = %+v, want invitation required", result)
	}
}
