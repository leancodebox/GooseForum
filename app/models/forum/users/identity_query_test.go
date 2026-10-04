package users

import (
	"errors"
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"strings"
	"testing"
	"time"
)

func TestNormalizedIdentityUsesUniqueIndexes(t *testing.T) {
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
	for _, row := range []EntityComplete{
		{Id: 1, Username: "  LegacyUser  ", Email: " Legacy@Example.com "},
		{Id: 2, Username: "other"},
		{Id: 3, Username: "another"},
	} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&EntityComplete{Username: "LEGACYUSER"}).Error; err == nil {
		t.Fatal("database accepted duplicate normalized username")
	}
	if err := db.Create(&EntityComplete{Username: "unique", Email: "LEGACY@example.com"}).Error; err == nil {
		t.Fatal("database accepted duplicate normalized email")
	}
	var nullEmails int64
	if err := db.Model(&EntityComplete{}).Where("email_normalized IS NULL").Count(&nullEmails).Error; err != nil || nullEmails != 2 {
		t.Fatalf("missing emails must use NULL: %d %v", nullEmails, err)
	}
	var rows []EntityComplete
	if err := db.Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0].Username != "LegacyUser" || rows[0].Email != "Legacy@Example.com" || rows[0].UsernameLower != "legacyuser" || rows[0].EmailNormalized != "legacy@example.com" || rows[1].EmailNormalized != "" {
		t.Fatalf("identity normalization: %+v", rows)
	}
	owner := identityTestUser(t)
	if err := CheckIdentityAvailable(strings.ToUpper(owner.Username), "", 0); !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("normalized username conflict: %v", err)
	}
	if err := CheckIdentityAvailable("", strings.ToUpper(owner.Email), 0); !errors.Is(err, ErrEmailExists) {
		t.Fatalf("normalized email conflict: %v", err)
	}
	for _, query := range []struct {
		sql   string
		args  []any
		index string
	}{
		{"SELECT id, username FROM users WHERE username_lower = ? AND id <> ? AND deleted_at IS NULL LIMIT 1", []any{"legacyuser", 0}, "ux_users_username_lower"},
		{"SELECT id, username FROM users WHERE email_normalized = ? AND id <> ? AND deleted_at IS NULL LIMIT 1", []any{"legacy@example.com", 0}, "ux_users_email_normalized"},
		{"SELECT id, username, restriction_status, restriction_until FROM users WHERE username_lower >= ? AND username_lower < ? AND (restriction_status IN (?, ?) OR restriction_status = '' OR restriction_until <= ?) AND deleted_at IS NULL ORDER BY username_lower LIMIT 8", []any{"leg", "leg~", RestrictionNormal, RestrictionSuspended, time.Now()}, "ux_users_username_lower"},
		{"SELECT id, username FROM users WHERE (id IN (?) OR username_lower IN (?)) AND deleted_at IS NULL", []any{1, "legacyuser"}, "ux_users_username_lower"},
	} {
		var plan []struct{ Detail string }
		if err := db.Raw("EXPLAIN QUERY PLAN "+query.sql, query.args...).Scan(&plan).Error; err != nil {
			t.Fatal(err)
		}
		details := ""
		for _, row := range plan {
			details += row.Detail + "\n"
		}
		if !strings.Contains(details, query.index) || strings.Contains(details, "TEMP B-TREE") || strings.Contains(details, "SCAN users") {
			t.Fatalf("inefficient plan: %s", details)
		}
	}
}

func TestIdentityConstraintsRejectConcurrentWrites(t *testing.T) {
	for _, identity := range []string{"username", "email"} {
		t.Run(identity, func(t *testing.T) {
			owner := identityTestUser(t)
			name := owner.Username + "-race"
			email := "race-" + owner.Email
			if err := CheckIdentityAvailable(name, email, 0); err != nil {
				t.Fatal(err)
			}
			start := make(chan struct{})
			results := make(chan error, 2)
			for i := range 2 {
				user := EntityComplete{Username: name, Email: email}
				if identity == "email" {
					user.Username = fmt.Sprintf("%s-%d", name, i)
				} else {
					user.Email = ""
				}
				go func() {
					<-start
					results <- Create(&user)
				}()
			}
			close(start)
			successes, conflicts := 0, 0
			want := ErrUsernameExists
			if identity == "email" {
				want = ErrEmailExists
			}
			for range 2 {
				err := <-results
				if err == nil {
					successes++
				} else if errors.Is(err, want) {
					conflicts++
				} else {
					t.Fatalf("unexpected write error: %v", err)
				}
			}
			builder().Unscoped().Where("username LIKE ?", name+"%").Delete(&EntityComplete{})
			if successes != 1 || conflicts != 1 {
				t.Fatalf("writes: success=%d conflicts=%d", successes, conflicts)
			}
		})
	}
}
