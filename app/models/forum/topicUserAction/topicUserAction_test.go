package topicUserAction

import (
	"reflect"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestTopicUserActionRepositoryParity(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate topic user action: %v", err)
	}
	conn.Where("1 = 1").Delete(&Entity{})

	if !SetLiked(1, 10, true) {
		t.Fatal("SetLiked(true) = false, want true")
	}
	got := GetByTopicId(1, 10)
	if got.Id == 0 || got.LikedAt == nil {
		t.Fatalf("GetByTopicId()=%#v, want liked row", got)
	}
	if !SetBookmarked(1, 10, true) || !SetWatched(1, 10, true) {
		t.Fatal("SetBookmarked/SetWatched true failed")
	}
	if !SetLiked(1, 10, false) {
		t.Fatal("SetLiked(false) = false, want changed")
	}
	got = GetByTopicId(1, 10)
	if got.LikedAt != nil || got.BookmarkedAt == nil || got.WatchedAt == nil {
		t.Fatalf("after updates row=%#v", got)
	}
	if SetLiked(1, 10, false) {
		t.Fatal("duplicate SetLiked(false) = true, want false")
	}

	t1 := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Minute)
	if !SetLikedAt(2, 10, &t1) || !SetLikedAt(2, 20, &t2) {
		t.Fatal("SetLikedAt() failed")
	}
	if !SetWatched(3, 10, true) || !SetWatched(4, 10, true) || !SetWatched(5, 10, true) {
		t.Fatal("SetWatched() failed")
	}

	watchers := ListActiveWatchUserIDsAfter(10, 2, []uint64{4}, 10)
	if !reflect.DeepEqual(watchers, []uint64{3, 5}) {
		t.Fatalf("ListActiveWatchUserIDsAfter()=%#v, want [3 5]", watchers)
	}

	refs, cursor := ListLikedTopicRefsBefore(2, "", 10)
	if len(refs) != 2 || refs[0].TopicID != 20 || refs[1].TopicID != 10 || cursor != "" {
		t.Fatalf("ListLikedTopicRefsBefore() refs=%#v cursor=%q", refs, cursor)
	}
}
