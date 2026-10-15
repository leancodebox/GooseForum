package users

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"gorm.io/gorm"
)

func TestTrustReadModelAndCAS(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	u := EntityComplete{Username: "test-trust", UsernameLower: "test-trust"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	old, err := getTrustState(db, u.Id)
	if err != nil {
		t.Fatal(err)
	}
	n := old
	n.TrustTopicIds = []uint64{99, 100}
	n.TrustLevel = trustlevel.Member
	n.TrustVisitDays = 2
	if ok, err := updateTrust(db, old, n); err != nil || !ok {
		t.Fatalf("update=%v %v", ok, err)
	}
	if ok, err := updateTrust(db, old, n); err != nil || ok {
		t.Fatalf("stale update=%v %v", ok, err)
	}
	got, err := getTrustState(db, u.Id)
	if err != nil || got.TrustVersion != 1 || len(got.TrustTopicIds) != 2 || got.TrustLevel != trustlevel.Member {
		t.Fatalf("read=%+v %v", got, err)
	}
	if err := db.Model(&EntityComplete{}).Where("id = ?", u.Id).Update("restriction_status", RestrictionSuspended).Error; err != nil {
		t.Fatal(err)
	}
	if ok, err := updateTrust(db, got, got); err != nil || ok {
		t.Fatal("concurrent restriction ignored")
	}
}
