package datamigration

import (
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func badgeStatsMigrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.AutoMigrate(&badgeStats.Entity{}, &badges.Entity{}, &userBadges.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestBackfillBadgeStatsIncludesSystemHistoryAndDeletedUsers(t *testing.T) {
	db := badgeStatsMigrationDB(t)
	now := time.Now()
	for _, user := range []users.EntityComplete{
		{Id: 1, Username: "current"},
		{Id: 2, Username: "deleted", DeletedAt: gorm.DeletedAt{Time: now, Valid: true}},
	} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&badges.Entity{Code: "unused", Name: "Unused"}).Error; err != nil {
		t.Fatal(err)
	}
	for _, grant := range []userBadges.Entity{
		{UserId: 1, BadgeCode: "system", GrantedAt: now},
		{UserId: 2, BadgeCode: "system", GrantedAt: now},
		{UserId: 1, BadgeCode: "revoked", GrantedAt: now, RevokedAt: &now},
		{UserId: 2, BadgeCode: "deleted_only", GrantedAt: now},
		{UserId: 999, BadgeCode: "orphaned_user", GrantedAt: now},
	} {
		if err := db.Create(&grant).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&badgeStats.Entity{BadgeCode: "stale", GrantedCount: 99, HasGrantHistory: true}).Error; err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := BackfillBadgeStatsWithDB(db, []string{"system", "system_empty"}); err != nil {
			t.Fatal(err)
		}
		var rows []badgeStats.Entity
		if err := db.Find(&rows).Error; err != nil {
			t.Fatal(err)
		}
		got := make(map[string]badgeStats.Entity)
		for _, row := range rows {
			got[row.BadgeCode] = row
		}
		for _, want := range []badgeStats.Entity{
			{BadgeCode: "system", GrantedCount: 1, HasGrantHistory: true},
			{BadgeCode: "system_empty"},
			{BadgeCode: "unused"},
			{BadgeCode: "revoked", HasGrantHistory: true},
			{BadgeCode: "deleted_only", HasGrantHistory: true},
			{BadgeCode: "orphaned_user", HasGrantHistory: true},
			{BadgeCode: "stale", HasGrantHistory: true},
		} {
			if row, exists := got[want.BadgeCode]; !exists || row != want {
				t.Fatalf("code=%s got=%+v exists=%v want=%+v", want.BadgeCode, row, exists, want)
			}
		}
	}
}

func TestBackfillBadgeStatsRollsBackPartialBatches(t *testing.T) {
	db := badgeStatsMigrationDB(t)
	if err := db.Create(&badgeStats.Entity{BadgeCode: "existing", GrantedCount: 7, HasGrantHistory: true}).Error; err != nil {
		t.Fatal(err)
	}
	const callback = "test:badge_stats_backfill_failure"
	if err := db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "badge_stats" {
			tx.AddError(errors.New("simulated stats write failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Create().Remove(callback) })
	if err := BackfillBadgeStatsWithDB(db, []string{"new_system"}); err == nil {
		t.Fatal("backfill ignored write failure")
	}
	var rows []badgeStats.Entity
	if err := db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].BadgeCode != "existing" || rows[0].GrantedCount != 7 || !rows[0].HasGrantHistory {
		t.Fatalf("failed backfill leaked writes: %+v", rows)
	}
}
