package articleUserAction

import (
	"testing"
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func setupTestDB(t *testing.T) {
	t.Helper()
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

func TestListLikedArticleRefsBefore(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	t1 := base.Add(time.Minute)
	t2 := base.Add(2 * time.Minute)
	t3 := base.Add(3 * time.Minute)
	SetLikedAt(1, 10, &t1)
	SetLikedAt(1, 20, &t2)
	SetLikedAt(1, 30, &t3)
	SetLikedAt(2, 40, &t3)
	SetBookmarkedAt(1, 50, &t3)

	got, nextCursor := ListLikedArticleRefsBefore(1, "", 2)
	if len(got) != 2 || got[0].ArticleID != 30 || got[1].ArticleID != 20 {
		t.Fatalf("ListLikedArticleRefsBefore() = %#v, want articles [30 20]", got)
	}
	if nextCursor == "" {
		t.Fatal("nextCursor is empty, want next cursor")
	}

	got, nextCursor = ListLikedArticleRefsBefore(1, nextCursor, 2)
	if len(got) != 1 || got[0].ArticleID != 10 {
		t.Fatalf("ListLikedArticleRefsBefore(next) = %#v, want article [10]", got)
	}
	if nextCursor != "" {
		t.Fatalf("nextCursor = %q, want empty", nextCursor)
	}
}
