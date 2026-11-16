package badgeStats

import (
	"errors"
	"math"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func statsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "stats.sqlite") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestAdjustPreservesHistoryAndRejectsMissingOrInvalidCounts(t *testing.T) {
	db := statsTestDB(t)
	if err := EnsureWithDB(db, "badge"); err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if _, err := LockWithDB(tx, "badge"); err != nil {
			return err
		}
		return AdjustWithDB(tx, "badge", 1, true)
	}); err != nil {
		t.Fatal(err)
	}
	if err := EnsureWithDB(db, "badge"); err != nil {
		t.Fatal(err)
	}
	var row Entity
	if err := db.First(&row, "badge_code = ?", "badge").Error; err != nil {
		t.Fatal(err)
	}
	if row.GrantedCount != 1 || !row.HasGrantHistory {
		t.Fatalf("ensure reset statistics: %+v", row)
	}
	if err := AdjustWithDB(db, "badge", -1, false); err != nil {
		t.Fatal(err)
	}
	if err := AdjustWithDB(db, "badge", -1, false); !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("underflow error=%v", err)
	}
	if err := db.First(&row, "badge_code = ?", "badge").Error; err != nil {
		t.Fatal(err)
	}
	if row.GrantedCount != 0 || !row.HasGrantHistory {
		t.Fatalf("revoke lost grant history: %+v", row)
	}
	if err := AdjustWithDB(db, "badge", 0, true); err != nil {
		t.Fatal(err)
	}
	if err := AdjustWithDB(db, "missing", 1, true); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing adjustment error=%v", err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		_, err := LockWithDB(tx, "missing")
		return err
	}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing lock error=%v", err)
	}
	if err := db.Model(&Entity{}).Where("badge_code = ?", "badge").Update("granted_count", int64(math.MaxInt64)).Error; err != nil {
		t.Fatal(err)
	}
	if err := AdjustWithDB(db, "badge", 1, false); !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("overflow error=%v", err)
	}
}

func TestSQLiteLockReadsAfterConcurrentWriterCommits(t *testing.T) {
	db := statsTestDB(t)
	if err := EnsureWithDB(db, "shared"); err != nil {
		t.Fatal(err)
	}
	first := db.Begin()
	if first.Error != nil {
		t.Fatal(first.Error)
	}
	t.Cleanup(func() { first.Rollback() })
	if _, err := LockWithDB(first, "shared"); err != nil {
		t.Fatal(err)
	}
	if err := AdjustWithDB(first, "shared", 1, true); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	var once sync.Once
	const callback = "test:stats_second_writer"
	if err := db.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == tableName {
			once.Do(func() { close(started) })
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Update().Remove(callback) })
	finished := make(chan error, 1)
	go func() {
		finished <- db.Transaction(func(tx *gorm.DB) error {
			row, err := LockWithDB(tx, "shared")
			if err != nil {
				return err
			}
			if row.GrantedCount != 1 {
				return errors.New("lock read stale count before acquiring write lock")
			}
			return AdjustWithDB(tx, "shared", 1, true)
		})
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("second writer did not start")
	}
	select {
	case err := <-finished:
		t.Fatalf("second writer finished before first committed: %v", err)
	case <-time.After(40 * time.Millisecond):
	}
	if err := first.Commit().Error; err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second writer did not complete")
	}
	var row Entity
	if err := db.First(&row, "badge_code = ?", "shared").Error; err != nil {
		t.Fatal(err)
	}
	if row.GrantedCount != 2 {
		t.Fatalf("concurrent count=%d", row.GrantedCount)
	}
}
