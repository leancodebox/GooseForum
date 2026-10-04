package mfaservice

import (
	"errors"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func TestAdminResetRevokesSessionsAndChallengesWithAudit(t *testing.T) {
	user, codes := enabledUser(t)
	cookie := newChallenge(t, user, "password")
	c, _ := testContext(cookie)
	if _, err := Login(c, codes[0]); err != nil {
		t.Fatal(err)
	}
	stale := newChallenge(t, user, "oauth")
	if err := adminReset(0, user.Id, ""); err == nil {
		t.Fatal("missing recovery reason accepted")
	}
	if err := adminReset(0, user.Id, "identity verified by operator"); err != nil {
		t.Fatal(err)
	}
	status, err := Status(user.Id)
	if err != nil || status["enabled"] != false || status["remainingCodes"] != int64(0) {
		t.Fatalf("remaining MFA: %v %v", status, err)
	}
	current, err := users.Get(user.Id)
	if err != nil || current.TokenVersion != user.TokenVersion+1 {
		t.Fatalf("version: %+v %v", current, err)
	}
	var active int64
	dbconnect.Connect().Model(&authsessions.Token{}).Where("user_id = ? AND revoked_at IS NULL", user.Id).Count(&active)
	if active != 0 {
		t.Fatal("session survived reset")
	}
	c, r := testContext(stale)
	if _, err = Login(c, codes[1]); err == nil || r.Header().Get("New-Token") != "" {
		t.Fatal("old challenge survived reset")
	}
	var log authsessions.Log
	if err = dbconnect.Connect().Where("user_id = ? AND action = ?", user.Id, "mfa_admin_reset").First(&log).Error; err != nil {
		t.Fatal(err)
	}
	if log.Reason != "identity verified by operator" || log.AuthMethod != "operator" || log.Result != "success" {
		t.Fatalf("missing audit reason: %+v", log)
	}
}

func TestAdminResetCanRetryAfterRecoveryCleanupFailure(t *testing.T) {
	user, _ := enabledUser(t)
	db := dbconnect.Connect()
	name := "test:mfa-reset-cleanup-failure"
	if err := db.Callback().Delete().Before("gorm:delete").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table == "user_mfa_recovery_codes" {
			tx.AddError(errors.New("cleanup unavailable"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Delete().Remove(name) })
	if err := adminReset(0, user.Id, "identity verified"); err == nil {
		t.Fatal("cleanup failure was ignored")
	}
	status, err := Status(user.Id)
	if err != nil || status["enabled"] != true {
		t.Fatal("factor removed before cleanup succeeded")
	}
	db.Callback().Delete().Remove(name)
	if err := adminReset(0, user.Id, "identity verified"); err != nil {
		t.Fatal(err)
	}
	status, err = Status(user.Id)
	if err != nil || status["enabled"] != false || status["remainingCodes"] != int64(0) {
		t.Fatalf("retry did not finish reset: %v %v", status, err)
	}
}

func TestAdminResetByRejectsMissingActorAndSelf(t *testing.T) {
	user, _ := enabledUser(t)
	for _, actor := range []uint64{0, user.Id} {
		if err := AdminResetBy(actor, user.Id, "identity verified"); err == nil {
			t.Fatal("invalid actor accepted")
		}
	}
	status, err := Status(user.Id)
	if err != nil || status["enabled"] != true {
		t.Fatal("rejected reset changed MFA")
	}
}
