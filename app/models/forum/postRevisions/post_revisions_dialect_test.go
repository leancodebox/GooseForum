package postRevisions_test

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestRevisionSchemaHasPortableUniqueConstraint(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&postRevisions.Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	indexes, err := db.Migrator().GetIndexes(&postRevisions.Entity{})
	if err != nil {
		t.Fatalf("get indexes: %v", err)
	}
	for _, index := range indexes {
		if index.Name() == "uniq_post_revision" {
			unique, ok := index.Unique()
			if !ok || !unique {
				t.Fatal("uniq_post_revision is not unique")
			}
			return
		}
	}
	t.Fatal("uniq_post_revision index is missing")
}

func TestRevisionPostLockSQLIsDialectAware(t *testing.T) {
	tests := []struct {
		name    string
		dialect gorm.Dialector
		want    string
		avoid   string
	}{
		{name: "sqlite", dialect: sqlite.Open(":memory:"), avoid: "FOR UPDATE"},
		{name: "mysql", dialect: mysql.New(mysql.Config{DSN: "", SkipInitializeWithVersion: true}), want: "FOR UPDATE"},
		{name: "postgres", dialect: postgres.New(postgres.Config{DSN: "", PreferSimpleProtocol: true}), want: "FOR UPDATE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(tt.dialect, &gorm.Config{DryRun: true, DisableAutomaticPing: true})
			if err != nil {
				t.Fatalf("open %s: %v", tt.name, err)
			}
			var row struct{ Id uint64 }
			statement := db.Table("posts").Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id").Where("id = ?", 1).First(&row).Statement.SQL.String()
			if tt.want != "" && !strings.Contains(strings.ToUpper(statement), tt.want) {
				t.Fatalf("lock SQL = %q, want %q", statement, tt.want)
			}
			if tt.avoid != "" && strings.Contains(strings.ToUpper(statement), tt.avoid) {
				t.Fatalf("lock SQL = %q, should omit %q", statement, tt.avoid)
			}
		})
	}
}
