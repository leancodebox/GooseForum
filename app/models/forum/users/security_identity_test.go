package users

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func identityTestUser(t *testing.T) EntityComplete {
	t.Helper()
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	activatedAt := time.Now()
	user := EntityComplete{Username: "identity-cas-" + suffix, Email: "identity-" + suffix + "@example.com", TokenVersion: 5, IsActivated: ActivationSuccess, ActivatedAt: &activatedAt}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&EntityComplete{}, user.Id) })
	return user
}

func readIdentityTestUser(t *testing.T, id uint64) EntityComplete {
	t.Helper()
	user, err := Get(id)
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func TestEmailAndUsernameChangesRejectStaleTokenVersion(t *testing.T) {
	user := identityTestUser(t)
	for _, update := range []func() error{
		func() error { return UpdateEmailByVersion(user.Id, user.TokenVersion-1, "other-"+user.Email) },
		func() error { return UpdateUsernameByVersion(user.Id, user.TokenVersion-1, user.Username+"-changed") },
	} {
		if err := update(); !errors.Is(err, ErrSecurityChanged) {
			t.Fatalf("stale update error=%v", err)
		}
	}
	actual := readIdentityTestUser(t, user.Id)
	if actual.Email != user.Email || actual.Username != user.Username || actual.TokenVersion != user.TokenVersion || actual.RequiresEmailVerification || actual.IsActivated != ActivationSuccess {
		t.Fatalf("stale update changed account: %+v", actual)
	}
}

func TestEmailAndUsernameChangesRejectCaseInsensitiveConflicts(t *testing.T) {
	owner, target := identityTestUser(t), identityTestUser(t)
	if err := UpdateEmailByVersion(target.Id, target.TokenVersion, strings.ToUpper(owner.Email)); !errors.Is(err, ErrEmailExists) {
		t.Fatalf("email conflict error=%v", err)
	}
	if err := UpdateUsernameByVersion(target.Id, target.TokenVersion, strings.ToUpper(owner.Username)); !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("username conflict error=%v", err)
	}
	actual := readIdentityTestUser(t, target.Id)
	if actual.Email != target.Email || actual.Username != target.Username || actual.TokenVersion != target.TokenVersion {
		t.Fatalf("conflict changed account: %+v", actual)
	}
}

func TestEmailChangeInvalidatesOldVersionAndRequiresVerification(t *testing.T) {
	user := identityTestUser(t)
	nextEmail := "changed-" + user.Email
	if err := UpdateEmailByVersion(user.Id, user.TokenVersion, nextEmail); err != nil {
		t.Fatal(err)
	}
	actual := readIdentityTestUser(t, user.Id)
	if actual.Email != nextEmail || actual.TokenVersion != user.TokenVersion+1 || actual.IsActivated != ActivationPending || !actual.RequiresEmailVerification || actual.ActivatedAt != nil {
		t.Fatalf("email update did not establish verification state: %+v", actual)
	}
	if !actual.NeedsEmailVerification(false) {
		t.Fatal("mandatory verification disappeared when global policy was disabled")
	}
	if err := UpdateEmailByVersion(user.Id, user.TokenVersion, user.Email); !errors.Is(err, ErrSecurityChanged) {
		t.Fatalf("old version restored previous email: %v", err)
	}
	if err := UpdateUsernameByVersion(user.Id, user.TokenVersion, user.Username+"-stale"); !errors.Is(err, ErrSecurityChanged) {
		t.Fatalf("old version renamed user after email change: %v", err)
	}
	if err := UpdateUsernameByVersion(user.Id, actual.TokenVersion, user.Username+"-current"); err != nil {
		t.Fatal(err)
	}
	latest := readIdentityTestUser(t, user.Id)
	if latest.Username != user.Username+"-current" || latest.Email != nextEmail || latest.TokenVersion != actual.TokenVersion {
		t.Fatalf("current username update lost security state: %+v", latest)
	}
}

func TestOperatorVerifiedEmailRecoveryInvalidatesSessionsWithoutVerificationLock(t *testing.T) {
	user := identityTestUser(t)
	if err := UpdateEmailByVersion(user.Id, user.TokenVersion, "unverified-"+user.Email); err != nil {
		t.Fatal(err)
	}
	if err := UpdateVerifiedEmailByVersion(user.Id, user.TokenVersion+1, "recovered-"+user.Email); err != nil {
		t.Fatal(err)
	}
	actual := readIdentityTestUser(t, user.Id)
	if actual.Email != "recovered-"+user.Email || actual.TokenVersion != user.TokenVersion+2 || actual.IsActivated != ActivationSuccess || actual.ActivatedAt == nil || actual.RequiresEmailVerification || actual.NeedsEmailVerification(true) {
		t.Fatalf("operator recovery left an unusable email state: %+v", actual)
	}
	if err := UpdateVerifiedEmailByVersion(user.Id, user.TokenVersion+1, user.Email); !errors.Is(err, ErrSecurityChanged) {
		t.Fatalf("stale operator recovery error=%v", err)
	}
}
