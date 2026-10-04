package users

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalAdministratorExistsMatchesRestrictionSemantics(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admins.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	for _, account := range []EntityComplete{{Id: 1, RoleId: 1, Username: "legacy-normal"}, {Id: 2, RoleId: 2, Username: "legacy-frozen", IsFrozen: 1}, {Id: 3, RoleId: 3, Username: "expired-ban", RestrictionStatus: RestrictionBanned, RestrictionUntil: &past, IsFrozen: 1}, {Id: 4, RoleId: 4, Username: "active-ban", RestrictionStatus: RestrictionBanned, RestrictionUntil: &future}, {Id: 5, RoleId: 5, Username: "explicit-normal", RestrictionStatus: RestrictionNormal, IsFrozen: 1}} {
		if err = db.Create(&account).Error; err != nil {
			t.Fatal(err)
		}
		got, err := hasNormalAdministrator(db, []uint64{account.RoleId}, 0, 0, now, false)
		want := account.EffectiveRestriction(now) == RestrictionNormal
		if err != nil || got != want {
			t.Fatalf("account=%+v exists=%v want=%v err=%v", account, got, want, err)
		}
	}
	for _, exclude := range [][2]uint64{{1, 0}, {0, 1}} {
		if found, err := hasNormalAdministrator(db, []uint64{1}, exclude[0], exclude[1], now, false); err != nil || found {
			t.Fatalf("administrator exclusion=%v found=%v err=%v", exclude, found, err)
		}
	}
	for _, account := range []EntityComplete{{Id: 6, RoleId: 6, Username: "verification-pending", RequiresEmailVerification: true}, {Id: 7, RoleId: 7, Username: "verification-complete", RequiresEmailVerification: true, IsActivated: ActivationSuccess}} {
		if err := db.Create(&account).Error; err != nil {
			t.Fatal(err)
		}
		for _, enabled := range []bool{false, true} {
			found, err := hasNormalAdministrator(db, []uint64{account.RoleId}, 0, 0, now, enabled)
			if err != nil || found != (account.IsActivated == ActivationSuccess) {
				t.Fatalf("verification enabled=%v account=%+v found=%v err=%v", enabled, account, found, err)
			}
		}
	}
	if found, err := hasNormalAdministrator(db, []uint64{1}, 0, 0, now, true); err != nil || found {
		t.Fatalf("legacy pending administrator with mandatory verification: found=%v err=%v", found, err)
	}
}
