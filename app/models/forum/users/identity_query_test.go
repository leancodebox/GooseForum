package users

import (
	"errors"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"strings"
	"testing"
)

func TestIdentityBackfillPreservesDuplicatesAndUsesIndexes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.AutoMigrate(&EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	for _, row := range []map[string]any{
		{"id": 1, "username": "LegacyUser", "email": "Legacy@Example.com"},
		{"id": 2, "username": "legacyuser", "email": "LEGACY@example.com"},
		{"id": 3, "username": "   ", "email": ""},
	} {
		if err := db.Table("users").Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if err := BackfillIdentityKeys(db); err != nil {
			t.Fatal(err)
		}
	}
	var rows []EntityComplete
	if err := db.Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0].Username != "LegacyUser" || rows[1].Email != "LEGACY@example.com" || rows[0].UsernameKey != "legacyuser" || rows[1].EmailKey != "legacy@example.com" {
		t.Fatalf("legacy accounts changed: %+v", rows)
	}
	if err := checkIdentityAvailable(db, "LEGACYUSER", "", 0); !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("legacy username conflict: %v", err)
	}
	if err := checkIdentityAvailable(db, "", "legacy@example.COM", 0); !errors.Is(err, ErrEmailExists) {
		t.Fatalf("legacy email conflict: %v", err)
	}
	for _, query := range []struct {
		sql   string
		args  []any
		index string
	}{
		{"SELECT id, username FROM users WHERE username_key = ? AND id <> ? AND deleted_at IS NULL LIMIT 1", []any{"legacyuser", 0}, "idx_users_username_key"},
		{"SELECT id, username FROM users WHERE email_key = ? AND id <> ? AND deleted_at IS NULL LIMIT 1", []any{"legacy@example.com", 0}, "idx_users_email_key"},
		{"SELECT id, username FROM users WHERE username_key >= ? AND username_key < ? AND deleted_at IS NULL ORDER BY username_key LIMIT 8", []any{"leg", "leg~"}, "idx_users_username_key"},
		{"SELECT id, username FROM users WHERE (id IN (?) OR username_key IN (?)) AND deleted_at IS NULL", []any{1, "legacyuser"}, "idx_users_username_key"},
	} {
		var plan []struct{ Detail string }
		if err := db.Raw("EXPLAIN QUERY PLAN "+query.sql, query.args...).Scan(&plan).Error; err != nil {
			t.Fatal(err)
		}
		details := ""
		for _, row := range plan {
			details += row.Detail + "\n"
		}
		if !strings.Contains(details, query.index) || strings.Contains(details, "TEMP B-TREE") {
			t.Fatalf("inefficient plan: %s", details)
		}
	}
}
