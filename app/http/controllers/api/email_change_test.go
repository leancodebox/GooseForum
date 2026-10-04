package api

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
)

func useMailSettings(t *testing.T, config pageConfig.MailSettingsConfig) {
	t.Helper()
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	original := pageConfig.GetByPageType(pageConfig.EmailSettings)
	t.Cleanup(func() {
		if original.Id != 0 {
			if err := pageConfig.SaveConfig(pageConfig.EmailSettings, original.Config); err != nil {
				t.Error(err)
			}
		} else if err := db.Where("page_type = ?", pageConfig.EmailSettings).Delete(&pageConfig.Entity{}).Error; err != nil {
			t.Error(err)
		}
		hotdataserve.ClearMailSettingsConfigCache()
	})
	if err := pageConfig.SaveConfig(pageConfig.EmailSettings, jsonopt.Encode(config)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearMailSettingsConfigCache()
}

func TestEditUserEmailSendsBeforeChangingIdentity(t *testing.T) {
	for _, name := range []string{"smtp-failure", "success"} {
		t.Run(name, func(t *testing.T) {
			useMailSettings(t, pageConfig.MailSettingsConfig{EnableMail: true, SmtpHost: "smtp.example.com", SmtpPort: 587, FromEmail: "forum@example.com"})
			db := dbconnect.Connect()
			if err := db.AutoMigrate(&users.EntityComplete{}, &usermfa.Factor{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
				t.Fatal(err)
			}
			user := users.MakeUser(fmt.Sprintf("email-send-%d", time.Now().UnixNano()), "password123", fmt.Sprintf("old-%d@example.com", time.Now().UnixNano()))
			user.IsActivated = users.ActivationSuccess
			if err := db.Create(user).Error; err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
				db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
				db.Unscoped().Delete(user)
			})
			c, issued := mfaAPIContext("", nil)
			if err := authsessionservice.Issue(c, user.Id, user.TokenVersion, authsessionservice.LoginDetails{Method: "password"}); err != nil {
				t.Fatal(err)
			}
			raw := issued.Header().Get("New-Token")
			if _, err := authsessionservice.Authenticate(c, raw, false); err != nil {
				t.Fatal(err)
			}
			newEmail := fmt.Sprintf("new-%d@example.com", user.Id)
			originalSend := sendEmailChangeVerification
			t.Cleanup(func() { sendEmailChangeVerification = originalSend })
			called := false
			sendEmailChangeVerification = func(prospective *users.EntityComplete) error {
				called = true
				stored, err := users.Get(user.Id)
				if err != nil || stored.Email != user.Email || stored.TokenVersion != user.TokenVersion {
					t.Fatalf("identity changed before SMTP: %+v %v", stored, err)
				}
				if prospective.Email != newEmail || prospective.IsActivated != users.ActivationPending || !prospective.RequiresEmailVerification {
					t.Fatalf("bad prospective recipient: %+v", prospective)
				}
				if name == "smtp-failure" {
					return errors.New("SMTP rejected email")
				}
				return nil
			}
			request, response := mfaAPIContext("", nil)
			result := EditUserEmail(component.BetterRequest[EditUserEmailReq]{UserId: user.Id, GinContext: request, Params: EditUserEmailReq{Email: newEmail}})
			if !called {
				t.Fatal("verification email was not attempted")
			}
			stored, err := users.Get(user.Id)
			if err != nil {
				t.Fatal(err)
			}
			if name == "smtp-failure" {
				if result.Data.MessageCode != component.MessageAuthActivationResendFailed || stored.Email != user.Email || stored.TokenVersion != user.TokenVersion || stored.IsActivated != users.ActivationSuccess || stored.RequiresEmailVerification {
					t.Fatalf("SMTP failure changed identity: %+v %+v", result, stored)
				}
				if response.Header().Get("Set-Cookie") != "" {
					t.Fatal("SMTP failure cleared cookie")
				}
				if _, err = authsessionservice.Authenticate(c, raw, false); err != nil {
					t.Fatalf("SMTP failure revoked existing session: %v", err)
				}
			} else {
				if result.Data.Code != component.SUCCESS || stored.Email != newEmail || stored.TokenVersion != user.TokenVersion+1 || stored.IsActivated != users.ActivationPending || !stored.RequiresEmailVerification {
					t.Fatalf("sent email did not update identity: %+v %+v", result, stored)
				}
				if response.Header().Get("Set-Cookie") == "" {
					t.Fatal("successful email change kept cookie")
				}
				if _, err = authsessionservice.Authenticate(c, raw, false); err == nil {
					t.Fatal("old session survived successful email change")
				}
			}
		})
	}
}

func TestEditUserEmailRejectsUnavailableMailWithoutChangingAccount(t *testing.T) {
	valid := pageConfig.MailSettingsConfig{EnableMail: true, SmtpHost: "smtp.example.com", SmtpPort: 587, FromEmail: "forum@example.com"}
	for _, name := range []string{"disabled", "missing-host", "invalid-port", "invalid-sender"} {
		t.Run(name, func(t *testing.T) {
			config := valid
			switch name {
			case "disabled":
				config.EnableMail = false
			case "missing-host":
				config.SmtpHost = " "
			case "invalid-port":
				config.SmtpPort = 0
			case "invalid-sender":
				config.FromEmail = "not an email"
			}
			useMailSettings(t, config)
			db := dbconnect.Connect()
			if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
				t.Fatal(err)
			}
			user := users.MakeUser(fmt.Sprintf("email-change-%d", time.Now().UnixNano()), "password123", fmt.Sprintf("original-%d@example.com", time.Now().UnixNano()))
			user.IsActivated = users.ActivationSuccess
			if err := db.Create(user).Error; err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
				db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
				db.Unscoped().Delete(user)
			})
			c, issued := mfaAPIContext("", nil)
			if err := authsessionservice.Issue(c, user.Id, user.TokenVersion, authsessionservice.LoginDetails{Method: "password"}); err != nil {
				t.Fatal(err)
			}
			raw := issued.Header().Get("New-Token")
			if _, err := authsessionservice.Authenticate(c, raw, false); err != nil {
				t.Fatal(err)
			}
			request, response := mfaAPIContext("", nil)
			result := EditUserEmail(component.BetterRequest[EditUserEmailReq]{
				UserId: user.Id, GinContext: request,
				Params: EditUserEmailReq{Email: fmt.Sprintf("replacement-%d@example.com", user.Id)},
			})
			if result.Data.Code != component.FAIL || result.Data.MessageCode != component.MessageAuthActivationResendFailed {
				t.Fatalf("missing mail accepted: %+v", result)
			}
			stored, err := users.Get(user.Id)
			if err != nil || stored.Email != user.Email || stored.TokenVersion != user.TokenVersion || stored.IsActivated != users.ActivationSuccess || stored.RequiresEmailVerification {
				t.Fatalf("account changed: %+v, %v", stored, err)
			}
			if response.Header().Get("Set-Cookie") != "" {
				t.Fatal("failed email change cleared session cookie")
			}
			if _, err := authsessionservice.Authenticate(c, raw, false); err != nil {
				t.Fatalf("existing session revoked: %v", err)
			}
		})
	}
}
