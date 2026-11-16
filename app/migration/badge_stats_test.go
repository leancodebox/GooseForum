package migration

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func badgeStatsVersionDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.AutoMigrate(&badgeStats.Entity{}, &badges.Entity{}, &userBadges.Entity{}, &users.EntityComplete{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := pageConfig.SyncMigrationVersionWithDB(db, 30); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestBadgeStatsMigrationDoesNotRecomputeAfterVersionAdvances(t *testing.T) {
	db := badgeStatsVersionDB(t)
	if err := migrateBadgeStatsWithDB(db, []string{"system_badge"}); err != nil {
		t.Fatal(err)
	}
	if version, err := pageConfig.GetMigrationVersionWithDB(db); err != nil || version != 31 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	if err := badgeStats.AdjustWithDB(db, "system_badge", 2, true); err != nil {
		t.Fatal(err)
	}
	// Any second aggregation would fail. A completed version must only retain
	// the values maintained by subsequent grant/revoke writes.
	if err := db.Migrator().DropTable(&userBadges.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := migrateBadgeStatsWithDB(db, []string{"system_badge"}); err != nil {
		t.Fatal(err)
	}
	var row badgeStats.Entity
	if err := db.First(&row, "badge_code = ?", "system_badge").Error; err != nil {
		t.Fatal(err)
	}
	if row.GrantedCount != 2 || !row.HasGrantHistory {
		t.Fatalf("restart recomputed statistics: %+v", row)
	}
}

func TestBadgeStatsMigrationRollsBackStatsAndVersionTogether(t *testing.T) {
	for _, failureTable := range []string{"badge_stats", "page_config"} {
		t.Run(failureTable, func(t *testing.T) {
			db := badgeStatsVersionDB(t)
			const callback = "test:badge_stats_migration_failure"
			if err := db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
				if tx.Statement.Table == failureTable {
					tx.AddError(errors.New("migration persistence failed"))
				}
			}); err != nil {
				t.Fatal(err)
			}
			if err := migrateBadgeStatsWithDB(db, []string{"system_badge"}); err == nil {
				t.Fatal("migration ignored persistence failure")
			}
			if version, err := pageConfig.GetMigrationVersionWithDB(db); err != nil || version != 30 {
				t.Fatalf("failed migration advanced version=%d err=%v", version, err)
			}
			var rows []badgeStats.Entity
			if err := db.Find(&rows).Error; err != nil || len(rows) != 0 {
				t.Fatalf("failed migration retained statistics=%+v err=%v", rows, err)
			}
			if err := db.Callback().Create().Remove(callback); err != nil {
				t.Fatal(err)
			}
			if err := migrateBadgeStatsWithDB(db, []string{"system_badge"}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
