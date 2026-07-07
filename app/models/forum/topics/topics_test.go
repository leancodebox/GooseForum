package topics

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"gorm.io/gorm"
)

func TestTopicAndPostSchemaMigrates(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := conn.AutoMigrate(&Entity{}, &posts.Entity{}); err != nil {
		t.Fatalf("migrate topic/post schema: %v", err)
	}

	if !conn.Migrator().HasTable("topics") {
		t.Fatal("topics table was not created")
	}
	if !conn.Migrator().HasTable("posts") {
		t.Fatal("posts table was not created")
	}

	for _, column := range []string{"title", "category_id", "first_post_id", "last_post_id", "post_count", "reply_count", "excerpt"} {
		if !conn.Migrator().HasColumn(&Entity{}, column) {
			t.Fatalf("topics.%s column was not created", column)
		}
	}

	for _, column := range []string{"topic_id", "post_no", "user_id", "content", "reply_to_post_id", "process_status"} {
		if !conn.Migrator().HasColumn(&posts.Entity{}, column) {
			t.Fatalf("posts.%s column was not created", column)
		}
	}

	if !conn.Migrator().HasIndex(&posts.Entity{}, "idx_posts_topic_no") {
		t.Fatal("posts idx_posts_topic_no index was not created")
	}
	if !conn.Migrator().HasIndex(&posts.Entity{}, "idx_posts_topic_created") {
		t.Fatal("posts idx_posts_topic_created index was not created")
	}
}
