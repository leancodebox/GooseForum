package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrationUsesCleanTopicEdgeModels(t *testing.T) {
	source, err := os.ReadFile("migration.go")
	if err != nil {
		t.Fatalf("read migration.go: %v", err)
	}

	text := string(source)
	for _, oldModel := range []string{
		"models/forum/articleCategory",
		"models/forum/articleCategoryRs",
		"models/forum/articleUserAction",
		"models/forum/articlesUserStat",
		"models/forum/articles",
		"models/forum/reply",
	} {
		if strings.Contains(text, oldModel) {
			t.Fatalf("migration still imports old edge model %q", oldModel)
		}
	}

	for _, cleanModel := range []string{
		"models/forum/accessGroupMembers",
		"models/forum/accessGroups",
		"models/forum/category",
		"models/forum/categoryGroupPermissions",
		"models/forum/migrationMapping",
		"models/forum/topicCategoryIndex",
		"models/forum/topicUserAction",
		"models/forum/topicUserStat",
	} {
		if !strings.Contains(text, cleanModel) {
			t.Fatalf("migration does not import clean edge model %q", cleanModel)
		}
	}
}

func TestActiveRuntimeDoesNotImportOldArticleReplyModels(t *testing.T) {
	roots := []string{
		"../http",
		"../service",
		"../models/hotdataserve",
	}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(source)
			for _, oldModel := range []string{
				"models/forum/articleCategory",
				"models/forum/articleCategoryRs",
				"models/forum/articleUserAction",
				"models/forum/articlesUserStat",
				"models/forum/articles",
				"models/forum/reply",
			} {
				if strings.Contains(text, oldModel) {
					t.Fatalf("%s imports old model %q", path, oldModel)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", root, err)
		}
	}
}

func TestStartupSchemaCreatesAllOIDCTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "startup.sqlite")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	// Use the actual startup registry, not Store.Migrate, and verify reruns too.
	for i := 0; i < 2; i++ {
		if err := db.AutoMigrate(defaultSchemaModels()...); err != nil {
			t.Fatal(err)
		}
		for _, model := range oidcProviderStore.Models() {
			if !db.Migrator().HasTable(model) {
				t.Fatalf("startup omitted OIDC table for %T", model)
			}
		}
		store, err := oidcProviderStore.New(db)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.ListSigningKeyHistory(t.Context()); err != nil {
			t.Fatalf("load signing key history after startup migration: %v", err)
		}
	}
}
