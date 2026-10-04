package authsessionservice

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func TestCachedSessionSeesRestrictionsAndExpiredBanDoesNotReviveToken(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: "restriction-session", TokenVersion: 5}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		InvalidateUser(user.Id)
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
		db.Unscoped().Delete(&user)
	})
	c, recorder := testContext()
	if err := Issue(c, user.Id, 5, LoginDetails{Method: "password"}); err != nil {
		t.Fatal(err)
	}
	var raw string
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "access_token" {
			raw = cookie.Value
		}
	}
	if _, err := Authenticate(c, raw, true); err != nil {
		t.Fatal(err)
	}
	// A separate process does not share this cache, so checking only cache invalidation is insufficient.
	if err := db.Model(&user).Updates(map[string]any{"restriction_status": users.RestrictionSuspended}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := Authenticate(c, raw, true); err != nil {
		t.Fatalf("suspension prevented read/security access: %v", err)
	}
	deadline := time.Now().Add(time.Minute)
	if err := db.Model(&user).Updates(map[string]any{"restriction_status": users.RestrictionBanned, "restriction_until": deadline, "token_version": gorm.Expr("token_version + 1")}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := Authenticate(c, raw, true); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("cached ban accepted: %v", err)
	}
	other, _ := testContext()
	if err := Issue(other, user.Id, 6, LoginDetails{}); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("banned login accepted: %v", err)
	}
	past := time.Now().Add(-time.Minute)
	if err := db.Model(&user).Update("restriction_until", past).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := Authenticate(c, raw, true); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expired ban revived old token: %v", err)
	}
	if err := Issue(other, user.Id, 6, LoginDetails{}); err != nil {
		t.Fatalf("expired ban prevented new login: %v", err)
	}
}

func TestUnverifiedIdentityRejectsCachedSessionsAndSessionIssuance(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: fmt.Sprintf("pending-session-%d", time.Now().UnixNano()), IsActivated: users.ActivationSuccess, RequiresEmailVerification: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		InvalidateUser(user.Id)
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
		db.Unscoped().Delete(&user)
	})
	c, recorder := testContext()
	if err := Issue(c, user.Id, user.TokenVersion, LoginDetails{Method: "password"}); err != nil {
		t.Fatal(err)
	}
	raw := recorder.Header().Get("New-Token")
	if _, err := Authenticate(c, raw, false); err != nil {
		t.Fatal(err)
	}
	// No version change or cache invalidation: authorization must read current verification state.
	if err := db.Model(&user).Update("is_activated", users.ActivationPending).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := Authenticate(c, raw, false); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("cached pending identity accepted: %v", err)
	}
	newContext, response := testContext()
	if err := Issue(newContext, user.Id, user.TokenVersion, LoginDetails{}); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("pending identity received ordinary session: %v", err)
	}
	consumed := false
	err := IssueVerified(newContext, user.Id, user.TokenVersion, LoginDetails{}, func() error {
		consumed = true
		return nil
	})
	if !errors.Is(err, ErrInvalidSession) || consumed || response.Header().Get("New-Token") != "" {
		t.Fatalf("pending challenge consumed factor or issued session: %v %v", consumed, err)
	}
}
