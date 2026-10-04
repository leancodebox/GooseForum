package accountrestrictionservice

import (
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accountrestrictions"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbconnect.Connect()
	models := []any{&users.EntityComplete{}, &role.Entity{}, &rolePermissionRs.Entity{}, &accountrestrictions.History{}, &oidcProviderStore.TokenEntity{}, &oidcProviderStore.AuthorizationCodeEntity{}, &pageConfig.Entity{}}
	if err := db.Migrator().DropTable(models...); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatal(err)
	}
	for _, r := range []role.Entity{{Id: 1, RoleName: "admin", Effective: 1}, {Id: 2, RoleName: "user manager", Effective: 1}, {Id: 3, RoleName: "site manager", Effective: 1}} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range []rolePermissionRs.Entity{{RoleId: 1, PermissionId: permission.Admin.Id(), Effective: 1}, {RoleId: 2, PermissionId: permission.UserManager.Id(), Effective: 1}, {RoleId: 3, PermissionId: permission.SiteManager.Id(), Effective: 1}} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, u := range []users.EntityComplete{{Id: 1, Username: "admin", RoleId: 1}, {Id: 2, Username: "otheradmin", RoleId: 1}, {Id: 3, Username: "manager", RoleId: 2}, {Id: 4, Username: "member", TokenVersion: 7}, {Id: 5, Username: "legacy", IsFrozen: 1}, {Id: 6, Username: "site-manager", RoleId: 3}} {
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestRestrictionVersionAuditExpiryAndProtection(t *testing.T) {
	db := testDB(t)
	now := time.Now()
	deadline := now.Add(time.Hour)
	change := Change{UserId: 4, Status: users.RestrictionSuspended, Until: &deadline, Reason: "spam", Note: "private"}
	if err := edit(1, change, now); err != nil {
		t.Fatal(err)
	}
	state, err := users.GetAccountState(4)
	if err != nil || state.TokenVersion != 7 || state.EffectiveRestriction(now) != users.RestrictionSuspended {
		t.Fatalf("suspension: %+v %v", state, err)
	}
	if state.EffectiveRestriction(deadline) != users.RestrictionNormal {
		t.Fatal("deadline does not expire")
	}
	change.Status = users.RestrictionBanned
	for _, token := range []oidcProviderStore.TokenEntity{{Hash: "opaqueaccess", Type: "access", UserID: "4", ClientID: "client", ExpiresAt: deadline}, {Hash: "opaquerefresh", Type: "refresh", UserID: "4", ClientID: "client", ExpiresAt: deadline}} {
		if err = db.Create(&token).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err = db.Create(&oidcProviderStore.AuthorizationCodeEntity{Hash: "pending-code", UserID: "4", ClientID: "client", RedirectURI: "https://app.example/callback", ExpiresAt: deadline, AuthTime: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err = edit(1, change, now); err != nil {
		t.Fatal(err)
	}
	state, _ = users.GetAccountState(4)
	if state.TokenVersion != 8 {
		t.Fatal("ban did not revoke old versions")
	}
	var external []oidcProviderStore.TokenEntity
	db.Find(&external)
	for _, token := range external {
		if token.RevokedAt == nil {
			t.Fatal("ban retained external token")
		}
	}
	var code oidcProviderStore.AuthorizationCodeEntity
	db.First(&code)
	if !code.Used {
		t.Fatal("ban retained authorization code")
	}
	change.Status = users.RestrictionNormal
	change.Reason = "appeal accepted"
	if err = edit(1, change, now); err != nil {
		t.Fatal(err)
	}
	state, _ = users.GetAccountState(4)
	if state.TokenVersion != 8 || state.RestrictionUntil != nil {
		t.Fatal("unban revived old token or retained deadline")
	}
	var history []accountrestrictions.History
	db.Order("id").Find(&history)
	if len(history) != 3 || history[0].Note != "private" || history[0].ActorId != 1 {
		t.Fatalf("audit: %+v", history)
	}
	for _, attempt := range []struct {
		actor  uint64
		change Change
	}{{1, Change{UserId: 1, Status: users.RestrictionBanned, Reason: "self", RoleId: 1}}, {3, Change{UserId: 2, Status: users.RestrictionBanned, Reason: "higher", RoleId: 1}}, {3, Change{UserId: 4, Status: users.RestrictionNormal, RoleId: 1}}} {
		if err = edit(attempt.actor, attempt.change, now); !errors.Is(err, ErrProtected) {
			t.Fatalf("protected operation: %v", err)
		}
	}
	legacy, _ := users.GetAccountState(5)
	if err = edit(3, Change{UserId: 4, Status: users.RestrictionNormal, RoleId: 3}, now); !errors.Is(err, ErrProtected) {
		t.Fatalf("user manager granted extra permissions: %v", err)
	}
	if err = edit(3, Change{UserId: 6, Status: users.RestrictionBanned, Reason: "higher permissions", RoleId: 3}, now); !errors.Is(err, ErrProtected) {
		t.Fatalf("user manager restricted extra permissions: %v", err)
	}
	if legacy.EffectiveRestriction(now) != users.RestrictionSuspended {
		t.Fatal("legacy semantics changed")
	}
}

func TestSequentialOperatorsCannotRemoveLastAdministrator(t *testing.T) {
	db := testDB(t)
	now := time.Now()
	results := make(chan error, 2)
	for _, id := range []uint64{1, 2} {
		results <- edit(0, Change{UserId: id, Status: users.RestrictionBanned, Reason: "test", RoleId: 1}, now)
	}
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful removals=%d; want one", success)
	}
	var admins []users.AccountState
	db.Model(&users.EntityComplete{}).Where("role_id = ?", 1).Find(&admins)
	active := 0
	for _, a := range admins {
		if a.EffectiveRestriction(now) == users.RestrictionNormal {
			active++
		}
	}
	if active != 1 {
		t.Fatalf("remaining admins=%d", active)
	}
}

func TestAuditFailureKeepsRestrictionAndRetryRepairsHistory(t *testing.T) {
	db := testDB(t)
	db.Callback().Create().Before("gorm:create").Register("test:history-failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "user_restriction_history" {
			tx.AddError(errors.New("audit failed"))
		}
	})
	if err := edit(1, Change{UserId: 4, Status: users.RestrictionBanned, Reason: "spam"}, time.Now()); err == nil {
		t.Fatal("audit failure ignored")
	}
	state, _ := users.GetAccountState(4)
	if state.TokenVersion != 8 || state.EffectiveRestriction(time.Now()) != users.RestrictionBanned {
		t.Fatal("failed audit should retain applied ban")
	}
	db.Callback().Create().Remove("test:history-failure")
	for range 2 {
		if err := edit(1, Change{UserId: 4, Status: users.RestrictionBanned, Reason: "spam"}, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	var history []accountrestrictions.History
	db.Find(&history)
	if len(history) != 1 {
		t.Fatalf("retry audit count=%d", len(history))
	}
}

func TestOIDCFailureCanBeRetriedWithoutChangingBan(t *testing.T) {
	db := testDB(t)
	now := time.Now()
	token := oidcProviderStore.TokenEntity{Hash: "retry-token", Type: "access", UserID: "4", ClientID: "client", ExpiresAt: now.Add(time.Hour)}
	if err := db.Create(&token).Error; err != nil {
		t.Fatal(err)
	}
	db.Callback().Update().Before("gorm:update").Register("test:revoke-failure", func(tx *gorm.DB) {
		if tx.Statement.Table == token.TableName() {
			tx.AddError(errors.New("revoke failed"))
		}
	})
	change := Change{UserId: 4, Status: users.RestrictionBanned, Reason: "spam"}
	if err := edit(1, change, now); err == nil {
		t.Fatal("revoke failure ignored")
	}
	state, _ := users.GetAccountState(4)
	if state.RestrictionStatus != users.RestrictionBanned || state.TokenVersion != 8 {
		t.Fatalf("ban not applied: %+v", state)
	}
	db.Callback().Update().Remove("test:revoke-failure")
	if err := edit(1, change, now); err != nil {
		t.Fatal(err)
	}
	db.First(&token)
	if token.RevokedAt == nil {
		t.Fatal("retry did not revoke token")
	}
}

func TestPendingVerificationCannotReplaceLastUsableAdministrator(t *testing.T) {
	db := testDB(t)
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", 1).Updates(map[string]any{"requires_email_verification": true, "is_activated": users.ActivationSuccess}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", 2).Update("requires_email_verification", true).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for _, change := range []Change{
		{UserId: 1, RoleId: 1, Status: users.RestrictionNormal, Validate: users.ActivationPending},
		{UserId: 1, RoleId: 1, Status: users.RestrictionBanned, Validate: users.ActivationSuccess, Reason: "restriction"},
		{UserId: 1, RoleId: 0, Status: users.RestrictionNormal, Validate: users.ActivationSuccess},
	} {
		if err := edit(0, change, now); !errors.Is(err, ErrProtected) {
			t.Fatalf("removed last usable administrator: %+v err=%v", change, err)
		}
	}
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", 2).Update("is_activated", users.ActivationSuccess).Error; err != nil {
		t.Fatal(err)
	}
	if err := edit(1, Change{UserId: 1, RoleId: 1, Status: users.RestrictionNormal, Validate: users.ActivationPending}, now); err != nil {
		t.Fatalf("second verified administrator should allow removing verification: %v", err)
	}
}
