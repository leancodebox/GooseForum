package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
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
	for range 2 {
		if err := db.AutoMigrate(defaultSchemaModels()...); err != nil {
			t.Fatal(err)
		}
		if !db.Migrator().HasTable(&topicrank.Entity{}) {
			t.Fatal("startup omitted ranking schedule table")
		}
		for _, model := range []any{&usermfa.Factor{}, &usermfa.RecoveryCode{}} {
			if !db.Migrator().HasTable(model) {
				t.Fatalf("startup omitted MFA table for %T", model)
			}
		}
		if db.Migrator().HasColumn(&topics.Entity{}, "next_rank_at") {
			t.Fatal("fresh startup recreated the old schedule column")
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

func TestStartupPreservesLegacyContentAndNotifications(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.sqlite")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	// Seed the old tables without the new version and deduplication columns.
	for _, statement := range []string{
		`CREATE TABLE posts (id integer PRIMARY KEY, topic_id integer NOT NULL, post_no integer NOT NULL, content text)`,
		`INSERT INTO posts (id, topic_id, post_no, content) VALUES (1, 1, 1, 'Literal @alice [mention user="123"]@alice[/mention]')`,
		`CREATE TABLE event_notification (id integer PRIMARY KEY, user_id bigint, topic_id integer NOT NULL DEFAULT 0)`,
		`INSERT INTO event_notification (id, user_id, topic_id) VALUES (1, 7, 1), (2, 7, 1)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if err := db.AutoMigrate(defaultSchemaModels()...); err != nil {
			t.Fatal(err)
		}
	}
	var post posts.Entity
	if err := db.First(&post, 1).Error; err != nil {
		t.Fatal(err)
	}
	if post.SourceVersion != 0 || post.Content != `Literal @alice [mention user="123"]@alice[/mention]` {
		t.Fatalf("startup changed legacy source: version=%d content=%q", post.SourceVersion, post.Content)
	}
	var notifications []eventNotification.Entity
	if err := db.Order("id").Find(&notifications).Error; err != nil {
		t.Fatal(err)
	}
	if len(notifications) != 2 || notifications[0].DedupeKey != nil || notifications[1].DedupeKey != nil {
		t.Fatalf("startup changed legacy notifications: %+v", notifications)
	}
	key := "post:1:user:7"
	first := eventNotification.Entity{UserId: 7, TopicId: 1, DedupeKey: &key}
	if err := db.Create(&first).Error; err != nil {
		t.Fatal(err)
	}
	duplicate := eventNotification.Entity{UserId: 7, TopicId: 1, DedupeKey: &key}
	if err := db.Create(&duplicate).Error; err == nil {
		t.Fatal("startup omitted the notification deduplication constraint")
	}
}
