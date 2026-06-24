package articleUserAction

import (
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	if err := db.Connect().AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate article user action: %v", err)
	}
	db.Connect().Where("1 = 1").Delete(&Entity{})
}

func TestSetStateStoresMultipleArticleUserFlagsInOneRow(t *testing.T) {
	setupTestDB(t)

	if !SetLiked(1, 10, true) {
		t.Fatal("SetLiked() = false, want true")
	}
	if !SetBookmarked(1, 10, true) {
		t.Fatal("SetBookmarked() = false, want true")
	}
	if !SetWatched(1, 10, true) {
		t.Fatal("SetWatched() = false, want true")
	}

	got := GetByArticleId(1, 10)
	if got.LikedAt == nil || got.BookmarkedAt == nil || got.WatchedAt == nil {
		t.Fatalf("state = likedAt:%v bookmarkedAt:%v watchedAt:%v, want all set", got.LikedAt, got.BookmarkedAt, got.WatchedAt)
	}
	SetLiked(1, 10, false)
	if got := GetByArticleId(1, 10); got.LikedAt != nil || got.BookmarkedAt == nil || got.WatchedAt == nil {
		t.Fatalf("after unlike state = likedAt:%v bookmarkedAt:%v watchedAt:%v, want only likedAt nil", got.LikedAt, got.BookmarkedAt, got.WatchedAt)
	}
	if SetLiked(1, 10, false) {
		t.Fatal("duplicate SetLiked(false) = true, want false")
	}
	if SetBookmarked(1, 10, true) {
		t.Fatal("duplicate SetBookmarked(true) = true, want false")
	}

	var count int64
	db.Connect().Model(&Entity{}).Where("user_id = ? AND article_id = ?", 1, 10).Count(&count)
	if count != 1 {
		t.Fatalf("row count = %d, want 1", count)
	}
}

func TestListActiveWatchUserIDsAfter(t *testing.T) {
	setupTestDB(t)

	SetWatched(1, 10, true)
	SetWatched(2, 10, true)
	SetWatched(3, 10, false)
	SetLiked(4, 10, true)
	SetWatched(5, 10, true)
	SetWatched(6, 20, true)

	got := ListActiveWatchUserIDsAfter(10, 1, []uint64{5}, 10)
	want := []uint64{2}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("ListActiveWatchUserIDsAfter() = %#v, want %#v", got, want)
	}
}
