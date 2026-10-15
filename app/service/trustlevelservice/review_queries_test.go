package trustlevelservice

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func TestAuthenticatedReviewQueryBudget(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	ClearReviewCache()
	t.Cleanup(ClearReviewCache)
	var queries atomic.Int32
	name := "trust-query-budget"
	if err := db.Callback().Query().Before("gorm:query").Register(name, func(*gorm.DB) { queries.Add(1) }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Query().Remove(name) })
	if err := ReviewAuthenticated(users.AccountState{Id: 999999, TrustLevel: trustlevel.Regular}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if queries.Load() != 0 {
		t.Fatal("stable member performed extra queries")
	}
	if err := ReviewAuthenticated(users.AccountState{Id: 999998, TrustLevelLocked: true}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if queries.Load() != 0 {
		t.Fatal("locked member performed extra queries")
	}
	state := users.AccountState{Id: 999997, CreatedAt: time.Now()}
	if err := ReviewAuthenticated(state, time.Now()); err != nil {
		t.Fatal(err)
	}
	queries.Store(0)
	if err := ReviewAuthenticated(state, time.Now()); err != nil {
		t.Fatal(err)
	}
	if queries.Load() != 0 {
		t.Fatal("repeat no-progress review performed queries")
	}
}
