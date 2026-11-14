package postRevisions_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"gorm.io/gorm"
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
