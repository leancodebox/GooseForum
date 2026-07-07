package category

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"gorm.io/gorm"
)

func TestCleanTopicEdgeModelsMigrate(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := conn.AutoMigrate(
		&Entity{},
		&topicCategoryIndex.Entity{},
		&topicUserStat.Entity{},
		&topicUserAction.Entity{},
	); err != nil {
		t.Fatalf("migrate clean topic edge models: %v", err)
	}

	for _, table := range []string{"category", "topic_category_index", "topic_user_stat", "topic_user_action"} {
		if !conn.Migrator().HasTable(table) {
			t.Fatalf("%s table was not created", table)
		}
	}

	for _, column := range []string{"topic_id", "category_id", "effective"} {
		if !conn.Migrator().HasColumn(&topicCategoryIndex.Entity{}, column) {
			t.Fatalf("topic_category_index.%s column was not created", column)
		}
	}

	for _, column := range []string{"topic_id", "user_id", "reply_count", "last_reply_at"} {
		if !conn.Migrator().HasColumn(&topicUserStat.Entity{}, column) {
			t.Fatalf("topic_user_stat.%s column was not created", column)
		}
	}

	for _, column := range []string{"topic_id", "user_id", "liked_at", "bookmarked_at", "watched_at"} {
		if !conn.Migrator().HasColumn(&topicUserAction.Entity{}, column) {
			t.Fatalf("topic_user_action.%s column was not created", column)
		}
	}

	if !conn.Migrator().HasIndex(&topicCategoryIndex.Entity{}, "idx_topic_category_effective") {
		t.Fatal("topic_category_index idx_topic_category_effective index was not created")
	}
	if !conn.Migrator().HasIndex(&topicUserAction.Entity{}, "uniq_user_topic_action") {
		t.Fatal("topic_user_action uniq_user_topic_action index was not created")
	}
	if !conn.Migrator().HasIndex(&topicUserStat.Entity{}, "idx_topic_user") {
		t.Fatal("topic_user_stat idx_topic_user index was not created")
	}
}
