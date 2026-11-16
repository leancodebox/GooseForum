package userBadges

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/sqlconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func concurrentStatsDB(t *testing.T) (*gorm.DB, uint64) {
	t.Helper()
	connection := sqlconnect.GetConnect(sqlconnect.Config{Connection: "sqlite", DbPath: filepath.Join(t.TempDir(), "grants.db"), MaxOpenConnections: 8, MaxIdleConnections: 8})
	if connection.Error != nil {
		t.Fatal(connection.Error)
	}
	db := connection.Connect
	underlying, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { underlying.Close() })
	if err := db.AutoMigrate(&Entity{}, &badgeStats.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: "member"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := badgeStats.EnsureWithDB(db, "award"); err != nil {
		t.Fatal(err)
	}
	return db, user.Id
}

func assertStats(t *testing.T, db *gorm.DB, count int64, history bool) {
	t.Helper()
	var stat badgeStats.Entity
	if err := db.Where("badge_code = ?", "award").First(&stat).Error; err != nil {
		t.Fatal(err)
	}
	if stat.GrantedCount != count || stat.HasGrantHistory != history {
		t.Fatalf("stats=%+v want count=%d history=%v", stat, count, history)
	}
}

func runConcurrentChanges(t *testing.T, change func(int) (bool, error), expected int) {
	t.Helper()
	var wait sync.WaitGroup
	results := make(chan bool, 12)
	failures := make(chan error, 12)
	for i := range 12 {
		wait.Go(func() { changed, err := change(i); results <- changed; failures <- err })
	}
	wait.Wait()
	close(results)
	close(failures)
	successes := 0
	for changed := range results {
		if changed {
			successes++
		}
	}
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if successes != expected {
		t.Fatalf("actual state changes=%d want=%d", successes, expected)
	}
}

func TestConcurrentMultiConnectionGrantRestoreAndRevokeCounts(t *testing.T) {
	db, userID := concurrentStatsDB(t)
	grant := func(int) (bool, error) { return grantWithDB(db, userID, "award", SourceManual, "thanks", 7, nil, true) }
	revoke := func(int) (bool, error) { return revokeWithDB(db, userID, "award") }
	runConcurrentChanges(t, grant, 1)
	assertStats(t, db, 1, true)
	runConcurrentChanges(t, revoke, 1)
	assertStats(t, db, 0, true)
	if changed, err := grantWithDB(db, userID, "award", SourceAuto, "retry", 0, nil, false); err != nil || changed {
		t.Fatalf("auto restored revoked award: %v %v", changed, err)
	}
	assertStats(t, db, 0, true)
	runConcurrentChanges(t, grant, 1)
	assertStats(t, db, 1, true)
	runConcurrentChanges(t, revoke, 1)
	assertStats(t, db, 0, true)
	var userIDs []uint64
	for i := range 12 {
		user := users.EntityComplete{Username: fmt.Sprintf("member-%d", i)}
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
		userIDs = append(userIDs, user.Id)
	}
	runConcurrentChanges(t, func(i int) (bool, error) {
		return grantWithDB(db, userIDs[i], "award", SourceAuto, "first", 0, nil, false)
	}, 12)
	assertStats(t, db, 12, true)
	runConcurrentChanges(t, func(i int) (bool, error) { return revokeWithDB(db, userIDs[i], "award") }, 12)
	assertStats(t, db, 0, true)
}

func TestGrantAndRevokeRollBackWhenStatsWriteFails(t *testing.T) {
	db, userID := concurrentStatsDB(t)
	failure := errors.New("statistics write unavailable")
	const callback = "test:reject_badge_stats_adjustment"
	install := func() {
		t.Helper()
		if err := db.Callback().Update().After("gorm:update").Register(callback, func(tx *gorm.DB) {
			if tx.Statement.Table != "badge_stats" {
				return
			}
			values, ok := tx.Statement.Dest.(map[string]any)
			if !ok {
				return
			}
			expression, ok := values["granted_count"].(clause.Expr)
			if ok && len(expression.Vars) > 0 {
				tx.AddError(failure)
			}
		}); err != nil {
			t.Fatal(err)
		}
	}
	remove := func() { db.Callback().Update().Remove(callback) }
	install()
	if changed, err := grantWithDB(db, userID, "award", SourceManual, "thanks", 7, nil, true); !errors.Is(err, failure) || changed {
		t.Fatalf("grant = %v %v", changed, err)
	}
	remove()
	assertStats(t, db, 0, false)
	var records []Entity
	if err := db.Where("badge_code = ?", "award").Find(&records).Error; err != nil || len(records) != 0 {
		t.Fatalf("grant was not rolled back: %+v %v", records, err)
	}
	if changed, err := grantWithDB(db, userID, "award", SourceManual, "thanks", 7, nil, true); err != nil || !changed {
		t.Fatalf("grant = %v %v", changed, err)
	}
	install()
	if changed, err := revokeWithDB(db, userID, "award"); !errors.Is(err, failure) || changed {
		t.Fatalf("revoke = %v %v", changed, err)
	}
	remove()
	assertStats(t, db, 1, true)
	var record Entity
	if err := db.Where("user_id = ? AND badge_code = ?", userID, "award").First(&record).Error; err != nil || record.RevokedAt != nil {
		t.Fatalf("revoke was not rolled back: %+v %v", record, err)
	}
}

func TestMissingStatsOrDeletedUsersCannotCreateAwards(t *testing.T) {
	db, userID := concurrentStatsDB(t)
	if changed, err := grantWithDB(db, userID, "missing", SourceManual, "thanks", 7, nil, true); !errors.Is(err, gorm.ErrRecordNotFound) || changed {
		t.Fatalf("missing stats = %v %v", changed, err)
	}
	if err := db.Delete(&users.EntityComplete{}, userID).Error; err != nil {
		t.Fatal(err)
	}
	if changed, err := grantWithDB(db, userID, "award", SourceManual, "thanks", 7, nil, true); !errors.Is(err, gorm.ErrRecordNotFound) || changed {
		t.Fatalf("deleted recipient = %v %v", changed, err)
	}
	assertStats(t, db, 0, false)
}
