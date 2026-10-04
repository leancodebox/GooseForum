package userservice

import (
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestProfileSaveCannotRestoreSecurityState(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	activatedAt := time.Now()
	stale := users.EntityComplete{Username: "stale-profile-" + suffix, Email: "old-" + suffix + "@example.com", Password: "old-hash", TokenVersion: 1, RoleId: 7, IsActivated: users.ActivationSuccess, ActivatedAt: &activatedAt}
	if err := db.Create(&stale).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&stale) })
	newUsername, newEmail := "current-profile-"+suffix, "current-"+suffix+"@example.com"
	until := time.Now().Add(time.Hour)
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", stale.Id).Updates(map[string]any{"username": newUsername, "email": newEmail, "is_activated": users.ActivationPending, "activated_at": nil, "requires_email_verification": true, "password": "new-hash", "token_version": 2, "role_id": 0, "restriction_status": users.RestrictionBanned, "restriction_reason": "spam", "restriction_until": until, "restriction_note": "moderator note", "is_frozen": 1}).Error; err != nil {
		t.Fatal(err)
	}
	stale.Bio = "profile update began before ban"
	if err := SaveUser(&stale); err != nil {
		t.Fatal(err)
	}
	if stale.Password != "new-hash" || stale.TokenVersion != 2 || stale.RoleId != 0 || stale.EffectiveRestriction(time.Now()) != users.RestrictionBanned {
		t.Fatalf("profile save overwrote security state: %+v", stale)
	}
	if stale.Username != newUsername || stale.Email != newEmail || stale.IsActivated != users.ActivationPending || !stale.RequiresEmailVerification || stale.ActivatedAt != nil || stale.RestrictionReason != "spam" || stale.RestrictionNote != "moderator note" || stale.RestrictionUntil == nil || !stale.RestrictionUntil.Equal(until) {
		t.Fatalf("profile save restored obsolete identity or verification fields: %+v", stale)
	}
	stored, err := users.Get(stale.Id)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Username != newUsername || stored.Email != newEmail || stored.Password != "new-hash" || stored.TokenVersion != 2 || stored.RoleId != 0 || !stored.RequiresEmailVerification || stored.ActivatedAt != nil || stored.RestrictionNote != "moderator note" {
		t.Fatalf("stored security state was overwritten: %+v", stored)
	}
	info, ok := GetUserInfo(stale.Id)
	if !ok || info.TokenVersion != 2 || info.RestrictionStatus != users.RestrictionBanned || info.Bio != stale.Bio {
		t.Fatalf("profile save refreshed stale cache: %+v", info)
	}
	if info.Username != newUsername || info.Email != newEmail || !info.RequiresEmailVerification || info.IsActivated != users.ActivationPending {
		t.Fatalf("identity cache contains obsolete fields: %+v", info)
	}
}
