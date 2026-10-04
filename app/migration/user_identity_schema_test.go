package migration

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestExistingUserIdentitySchemaUpgrade(t *testing.T) {
	db := dbconnect.Connect()
	t.Cleanup(func() { db.Migrator().DropTable(&users.EntityComplete{}) })
	for _, sql := range []string{
		"CREATE TABLE users (id integer PRIMARY KEY, username varchar(64) NOT NULL, email varchar(128), deleted_at datetime)",
		"INSERT INTO users VALUES (1, 'Alice', 'Alice@Example.COM', NULL), (2, '用户甲', NULL, NULL), (3, 'Deleted', '', '2026-01-01')",
		"CREATE UNIQUE INDEX ux_users_email ON users(email)",
		"CREATE UNIQUE INDEX ux_users_username ON users(username)",
		"WITH RECURSIVE seq(id) AS (SELECT 4 UNION ALL SELECT id + 1 FROM seq WHERE id < 603) INSERT INTO users SELECT id, 'User' || id, NULL, NULL FROM seq",
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	// Also cover retry after columns were added but population was interrupted.
	if err := db.Exec("ALTER TABLE users ADD COLUMN username_lower varchar(64)").Error; err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := users.PrepareIdentitySchema(); err != nil {
			t.Fatal(err)
		}
		if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
			t.Fatal(err)
		}
	}
	var rows []users.EntityComplete
	if err := db.Unscoped().Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 603 || rows[0].Username != "Alice" || rows[0].UsernameLower != "alice" || rows[0].EmailNormalized != "alice@example.com" || rows[1].UsernameLower != "用户甲" || rows[1].Email != "" || rows[2].UsernameLower != "deleted" || rows[602].UsernameLower != "user603" {
		t.Fatalf("incorrect upgraded identities: %+v", rows)
	}
	var missingEmails int64
	if err := db.Model(&users.EntityComplete{}).Unscoped().Where("email_normalized IS NULL").Count(&missingEmails).Error; err != nil || missingEmails != 602 {
		t.Fatalf("missing emails=%d error=%v", missingEmails, err)
	}
	if err := db.Create(&users.EntityComplete{Username: "ALICE"}).Error; err == nil {
		t.Fatal("upgraded database accepted duplicate normalized username")
	}
}
