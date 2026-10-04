package mfaservice

import (
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

func failMFAWrite(t *testing.T, operation, table string) func() {
	t.Helper()
	callback := dbconnect.Connect().Callback().Create()
	if operation == "delete" {
		callback = dbconnect.Connect().Callback().Delete()
	}
	name := "test:mfa-change-failure"
	if err := callback.Before("gorm:"+operation).Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			tx.AddError(errors.New("temporary database failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	remove := func() { callback.Remove(name) }
	t.Cleanup(remove)
	return remove
}

func TestEnableRetriesAfterFactorWriteFailure(t *testing.T) {
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
	remove := failMFAWrite(t, "create", "user_mfa")
	if _, err := Change(c, user.Id, "password123", code, "enable"); err == nil {
		t.Fatal("factor write failure ignored")
	}
	if enabled, err := usermfa.HasFactor(user.Id); err != nil || enabled {
		t.Fatalf("failed enable activated MFA: %v %v", enabled, err)
	}
	remove()
	codes, err := Change(c, user.Id, "password123", code, "enable")
	if err != nil || len(codes) != 10 {
		t.Fatalf("enable retry: %d %v", len(codes), err)
	}
	if count, err := usermfa.CountUnusedRecoveryCodes(user.Id); err != nil || count != 10 {
		t.Fatalf("orphan recovery codes survived retry: %d %v", count, err)
	}
}

func TestRegenerateFailurePreservesOldCodes(t *testing.T) {
	for _, operation := range []string{"create", "delete"} {
		t.Run(operation, func(t *testing.T) {
			user, codes := enabledUser(t)
			c, _ := testContext(nil)
			remove := failMFAWrite(t, operation, "user_mfa_recovery_codes")
			if _, err := Change(c, user.Id, "password123", codes[0], "regenerate"); err == nil {
				t.Fatal("recovery replacement failure ignored")
			}
			remove()
			replacement, err := Change(c, user.Id, "password123", codes[1], "regenerate")
			if err != nil || len(replacement) != 10 {
				t.Fatalf("old code cannot retry replacement: %d %v", len(replacement), err)
			}
			if count, err := usermfa.CountUnusedRecoveryCodes(user.Id); err != nil || count != 10 {
				t.Fatalf("replacement left extra codes: %d %v", count, err)
			}
		})
	}
}

func TestConcurrentRecoveryChangesConsumeCodeOnce(t *testing.T) {
	user, codes := enabledUser(t)
	results := make(chan error, 2)
	for range 2 {
		c, _ := testContext(nil)
		go func() {
			_, err := Change(c, user.Id, "password123", codes[0], "regenerate")
			results <- err
		}()
	}
	successes := 0
	for range 2 {
		if <-results == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent changes succeeded %d times", successes)
	}
}

func TestDisableRetriesAfterFactorDeleteFailure(t *testing.T) {
	user, codes := enabledUser(t)
	c, _ := testContext(nil)
	remove := failMFAWrite(t, "delete", "user_mfa")
	if _, err := Change(c, user.Id, "password123", codes[0], "disable"); err == nil {
		t.Fatal("factor deletion failure ignored")
	}
	factor, err := usermfa.GetFactor(user.Id)
	if err != nil {
		t.Fatal("failed disable removed factor")
	}
	secret, err := open(user.Id, factor.Secret)
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(secret, time.Now().Add(30*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	remove()
	if _, err := Change(c, user.Id, "password123", code, "disable"); err != nil {
		t.Fatalf("authenticator cannot finish disable: %v", err)
	}
	if enabled, err := usermfa.HasFactor(user.Id); err != nil || enabled {
		t.Fatalf("disable retry left factor: %v %v", enabled, err)
	}
}
