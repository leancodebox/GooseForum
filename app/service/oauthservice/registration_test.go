package oauthservice

import (
	"errors"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/registrationservice"
	"github.com/markbates/goth"
)

func withRegistrationConfig(t *testing.T, raw string) {
	t.Helper()
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&users.EntityComplete{}, &userOAuth.Entity{}, &userStatistics.Entity{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	previous := pageConfig.GetByPageType(pageConfig.SecuritySettings)
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, raw); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
	t.Cleanup(func() {
		if previous.Id != 0 {
			_ = pageConfig.SaveConfig(pageConfig.SecuritySettings, previous.Config)
		} else {
			conn.Where("page_type = ?", pageConfig.SecuritySettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearSecuritySettingsConfigCache()
		conn.Where("provider = ?", "registration-test").Delete(&userOAuth.Entity{})
		conn.Where("username LIKE ?", "registration-test-%").Delete(&users.EntityComplete{})
	})
}

func TestRegistrationClosedKeepsExistingOAuthLogin(t *testing.T) {
	withRegistrationConfig(t, `{"enableSignup":false}`)
	conn := dbconnect.Connect()
	u := users.MakeUser("registration-test-existing", "password", "existing@example.com")
	if err := conn.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&userOAuth.Entity{UserId: u.Id, Provider: "registration-test", ProviderUid: "existing"}).Error; err != nil {
		t.Fatal(err)
	}
	found, err := ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: "existing"}, "192.0.2.1")
	if err != nil || found.Id != u.Id {
		t.Fatalf("existing login: user=%+v error=%v", found, err)
	}
	_, err = ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: "new", Email: "new@example.com"}, "192.0.2.1")
	if !errors.Is(err, registrationservice.ErrClosed) {
		t.Fatalf("new account: %v", err)
	}
}

func TestRegistrationClosedRepairsPendingAccountWithExistingBinding(t *testing.T) {
	withRegistrationConfig(t, `{"enableSignup":false}`)
	conn := dbconnect.Connect()
	u := users.MakeUser("registration-test-repair", "password", "repair@example.com")
	key := users.OAuthRegistrationKey("registration-test", "repair")
	u.OAuthRegistrationKey = &key
	u.IsActivated = users.ActivationSuccess
	if err := conn.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&userOAuth.Entity{UserId: u.Id, Provider: "registration-test", ProviderUid: "repair"}).Error; err != nil {
		t.Fatal(err)
	}
	found, err := ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: "repair", Name: "Recovered profile"}, "192.0.2.11")
	if err != nil || found.Id != u.Id || found.Nickname != "Recovered profile" {
		t.Fatalf("repair: user=%+v error=%v", found, err)
	}
	stored, err := users.Get(u.Id)
	if err != nil || stored.OAuthRegistrationKey != nil {
		t.Fatalf("repair retained pending marker after profile save: user=%+v error=%v", stored, err)
	}
}

func TestOAuthNewAccountCannotBypassRequiredDomain(t *testing.T) {
	withRegistrationConfig(t, `{"enableSignup":true,"allowedDomains":["example.com"]}`)
	for _, email := range []string{"", "user@sub.example.com", "user@other.test"} {
		_, err := ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: "blocked", Email: email}, "192.0.2.2")
		if !errors.Is(err, registrationservice.ErrDomain) && !errors.Is(err, registrationservice.ErrEmailRequired) {
			t.Fatalf("email=%q error=%v", email, err)
		}
	}
}

func TestOAuthEmailConflictDoesNotAutomaticallyBind(t *testing.T) {
	withRegistrationConfig(t, `{"enableSignup":true,"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`)
	conn := dbconnect.Connect()
	u := users.MakeUser("registration-test-emailowner", "password", "owner@example.com")
	if err := conn.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	_, err := ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: "foreign", NickName: "registration-test-foreign", Email: "OWNER@example.com"}, "192.0.2.3")
	if !errors.Is(err, users.ErrEmailExists) {
		t.Fatalf("email conflict: %v", err)
	}
	if userOAuth.GetByProviderAndUID("registration-test", "foreign") != nil {
		t.Fatal("foreign provider was bound to existing email owner")
	}
}

func TestOAuthDomainPolicyRequiresVerifiedMailbox(t *testing.T) {
	withRegistrationConfig(t, `{"enableSignup":true,"enableEmailVerification":false,"allowedDomains":["example.com"],"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`)
	withOAuthRegistrationMail(t, `{"enableMail":true,"smtpHost":"smtp.example.com","smtpPort":587,"fromEmail":"forum@example.com"}`)
	for _, verified := range []bool{false, true} {
		uid, email, name := "unverified", "unverified@example.com", "registration-test-unverified"
		if verified {
			uid, email, name = "verified", "verified@example.com", "registration-test-verified"
		}
		user, err := ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: uid, NickName: name, Email: email, RawData: map[string]any{"email_verified": verified}}, "192.0.2.4")
		if err != nil {
			t.Fatal(err)
		}
		want := int8(users.ActivationPending)
		if verified {
			want = users.ActivationSuccess
		}
		if user.IsActivated != want {
			t.Fatalf("verified=%v state=%v", verified, user.IsActivated)
		}
	}
}

func withOAuthRegistrationMail(t *testing.T, raw string) {
	t.Helper()
	previous := pageConfig.GetByPageType(pageConfig.EmailSettings)
	if err := pageConfig.SaveConfig(pageConfig.EmailSettings, raw); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearMailSettingsConfigCache()
	t.Cleanup(func() {
		if previous.Id != 0 {
			_ = pageConfig.SaveConfig(pageConfig.EmailSettings, previous.Config)
		} else {
			dbconnect.Connect().Where("page_type = ?", pageConfig.EmailSettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearMailSettingsConfigCache()
	})
}

func TestOAuthRegistrationWithUnverifiedEmailRequiresConfiguredMail(t *testing.T) {
	withRegistrationConfig(t, `{"enableSignup":true,"enableEmailVerification":true,"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`)
	withOAuthRegistrationMail(t, `{"enableMail":false}`)
	_, err := ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: "no-mail", NickName: "registration-test-no-mail", Email: "no-mail@example.com"}, "192.0.2.200")
	if !errors.Is(err, registrationservice.ErrMailUnavailable) {
		t.Fatalf("missing mail error: %v", err)
	}
	if users.ExistEmail("no-mail@example.com") || userOAuth.GetByProviderAndUID("registration-test", "no-mail") != nil {
		t.Fatal("registration wrote an unusable account")
	}
	verified, err := ProcessOAuthCallback(goth.User{Provider: "registration-test", UserID: "trusted-mail", NickName: "registration-test-trusted-mail", Email: "trusted-mail@example.com", RawData: map[string]any{"email_verified": true}}, "192.0.2.200")
	if err != nil || verified.IsActivated != users.ActivationSuccess {
		t.Fatalf("verified provider needlessly blocked: %v", err)
	}
}
