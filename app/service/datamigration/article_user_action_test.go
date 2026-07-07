package datamigration

import (
	"testing"
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
)

func TestBackfillArticleUserActionMergesLegacyRows(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := db.Connect()
	createLegacyArticleActionTables(t)

	actedAt := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	for _, table := range []string{"article_like", "article_bookmark", "article_watch"} {
		conn.Table(table).Create(map[string]any{"user_id": 1, "article_id": 10, "status": 1, "updated_at": actedAt})
	}
	conn.Table("article_watch").Create(map[string]any{"user_id": 2, "article_id": 10, "status": 0, "updated_at": actedAt})

	result := BackfillArticleUserAction()
	if result.Failed != 0 {
		t.Fatalf("BackfillArticleUserAction() failed = %d, want 0", result.Failed)
	}

	var got topicUserAction.Entity
	conn.Where("user_id = ? AND topic_id = ?", 1, 10).First(&got)
	if got.LikedAt == nil || got.BookmarkedAt == nil || got.WatchedAt == nil {
		t.Fatalf("state = likedAt:%v bookmarkedAt:%v watchedAt:%v, want all set", got.LikedAt, got.BookmarkedAt, got.WatchedAt)
	}
	if !got.LikedAt.Equal(actedAt) || !got.BookmarkedAt.Equal(actedAt) || !got.WatchedAt.Equal(actedAt) {
		t.Fatalf("state times = %v %v %v, want %v", got.LikedAt, got.BookmarkedAt, got.WatchedAt, actedAt)
	}
	var inactive topicUserAction.Entity
	conn.Where("user_id = ? AND topic_id = ?", 2, 10).First(&inactive)
	if inactive.Id != 0 {
		t.Fatalf("inactive watch backfilled id=%d, want 0", inactive.Id)
	}
}

func TestBackfillArticleUserActionSkipsMissingLegacyTables(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")

	result := BackfillArticleUserAction()
	if result.Failed != 0 {
		t.Fatalf("BackfillArticleUserAction() failed = %d, want 0", result.Failed)
	}
}

func createLegacyArticleActionTables(t *testing.T) {
	t.Helper()
	conn := db.Connect()
	for _, table := range []string{"article_like", "article_bookmark", "article_watch"} {
		if err := conn.Exec(`DROP TABLE IF EXISTS ` + table).Error; err != nil {
			t.Fatalf("drop %s: %v", table, err)
		}
		if err := conn.Exec(`CREATE TABLE ` + table + ` (
			id integer primary key autoincrement,
			user_id integer not null default 0,
			article_id integer not null default 0,
			status integer not null default 1,
			updated_at datetime
		)`).Error; err != nil {
			t.Fatalf("create %s: %v", table, err)
		}
	}
	db.Connect().Where("1 = 1").Delete(&topicUserAction.Entity{})
}
