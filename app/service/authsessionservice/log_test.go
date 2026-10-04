package authsessionservice

import (
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func TestAuditFailureDoesNotRejectCreatedSession(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: "audit-test-" + time.Now().Format("150405.000000000"), TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{}); db.Unscoped().Delete(&user) })
	const callback = "test:reject_auth_audit"
	if err := db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "user_auth_token_logs" {
			tx.AddError(errors.New("audit storage unavailable"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Create().Remove(callback) })
	c, recorder := testContext()
	if err := Issue(c, user.Id, user.TokenVersion, LoginDetails{Method: "password"}); err != nil {
		t.Fatalf("audit rejected login: %v", err)
	}
	if recorder.Header().Get("New-Token") == "" {
		t.Fatal("session was not issued")
	}
}
