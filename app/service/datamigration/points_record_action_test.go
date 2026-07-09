package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigratePointsRecordActionBackfillsLegacyReason(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.Exec(`CREATE TABLE points_record (
		id integer primary key autoincrement,
		user_id integer not null default 0,
		change_reason text,
		action text not null default '',
		points_change integer not null default 0
	)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	if err := conn.Exec(`INSERT INTO points_record (user_id, change_reason) VALUES (1, '初始化'), (2, '旧原因')`).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	result := MigratePointsRecordActionWithDB(conn)
	if result.Failed > 0 {
		t.Fatalf("migration failed: %+v", result)
	}
	if result.Backfilled != 2 {
		t.Fatalf("Backfilled = %d, want 2", result.Backfilled)
	}
	if !result.ChangeReasonColumnDropped {
		t.Fatal("ChangeReasonColumnDropped = false, want true")
	}
	if conn.Migrator().HasColumn("points_record", "change_reason") {
		t.Fatal("change_reason column still exists")
	}

	var rows []struct {
		UserID uint64 `gorm:"column:user_id"`
		Action string
	}
	if err := conn.Table("points_record").Order("user_id").Find(&rows).Error; err != nil {
		t.Fatalf("read rows: %v", err)
	}
	if rows[0].Action != "init" {
		t.Fatalf("first action = %q, want init", rows[0].Action)
	}
	if rows[1].Action != "unknown" {
		t.Fatalf("second action = %q, want unknown", rows[1].Action)
	}
}

func TestMigratePointsRecordActionSkipsFreshSchema(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.Exec(`CREATE TABLE points_record (
		id integer primary key autoincrement,
		user_id integer not null default 0,
		action text not null default '',
		points_change integer not null default 0
	)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	result := MigratePointsRecordActionWithDB(conn)
	if result.Failed > 0 || result.Backfilled != 0 {
		t.Fatalf("fresh schema migration result = %+v, want no-op", result)
	}
}
