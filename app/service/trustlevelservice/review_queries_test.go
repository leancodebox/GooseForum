package trustlevelservice

import (
	"sync/atomic"
	"testing"
	"time"

	"encoding/json"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"gorm.io/gorm"
)

func TestReplyAnalysisLoadsAccountOnce(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &pageConfig.Entity{}, &posts.Entity{}); err != nil {
		t.Fatal(err)
	}
	c := pageConfig.DefaultTrustSettings()
	raw, _ := json.Marshal(c)
	previous := pageConfig.GetByPageType(pageConfig.TrustSettings)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	t.Cleanup(func() {
		if previous.Id > 0 {
			_ = pageConfig.SaveConfig(pageConfig.TrustSettings, previous.Config)
		} else {
			db.Where("page_type = ?", pageConfig.TrustSettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearTrustSettingsCache()
	})
	u := users.EntityComplete{Username: "reply-budget", UsernameLower: "reply-budget"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	var reads atomic.Int32
	name := "reply-policy-budget"
	if err := db.Callback().Query().Before("gorm:query").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table == "users" {
			reads.Add(1)
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Query().Remove(name) })
	analysis, err := AnalyzePost(u.Id, "hello ![a](/a.png)", 0, true)
	if err != nil || analysis.RenderedHTML == "" {
		t.Fatalf("analysis=%+v error=%v", analysis, err)
	}
	if reads.Load() != 1 {
		t.Fatalf("account queries=%d, want 1", reads.Load())
	}
}

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
