package topicrankservice

import (
	"context"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

func rankDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbconnect.Connect()
	if err := db.Migrator().DropTable(&topicrank.Entity{}, &topics.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&topics.Entity{}, &topicrank.Entity{}); err != nil {
		t.Fatal(err)
	}
	return db
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func readTopic(t *testing.T, db *gorm.DB, id uint64) topics.Entity {
	t.Helper()
	var row topics.Entity
	must(t, db.First(&row, id).Error)
	return row
}

func readSchedule(t *testing.T, id uint64) topicrank.Entity {
	t.Helper()
	entry, err := topicrank.Get(context.Background(), id)
	must(t, err)
	return entry
}

func TestFreshnessBoundariesAndCaps(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		age  time.Duration
		want int64
	}{{0, 30}, {6 * time.Hour, 24}, {12 * time.Hour, 18}, {24 * time.Hour, 12}, {48 * time.Hour, 8}, {72 * time.Hour, 4}, {7 * 24 * time.Hour, 0}} {
		got, next := Score(Signals{}, now.Add(-tc.age), now)
		if got != tc.want*1000 {
			t.Fatalf("age %v score %v want %v", tc.age, got, tc.want)
		}
		if next != nil && !next.After(now) {
			t.Fatal("non-future boundary")
		}
	}
	got, next := Score(Signals{Likes: 99999, Replies: 99999}, now.Add(-30*24*time.Hour), now)
	if got != 20000 || next != nil {
		t.Fatalf("dormant score=%v next=%v", got, next)
	}
}

func TestDirtyMarkerAndStaleSave(t *testing.T) {
	db := rankDB(t)
	topic := topics.Entity{Id: 1, Status: 1}
	must(t, saveAndHandleRank(db, &topic))
	stale := readTopic(t, db, 1)
	must(t, db.Table("topics").Where("id = 1").Updates(map[string]any{"rank_score": 42}).Error)
	stale.Title = "edited"
	must(t, saveAndHandleRank(db, &stale))
	got := readTopic(t, db, 1)
	if got.RankScore != 42 || readSchedule(t, 1).NextRunAt == nil {
		t.Fatalf("stale save overwrote ranking: %+v", got)
	}

}

func TestBackfillAndDraftFirstPublication(t *testing.T) {
	db := rankDB(t)
	old := time.Now().Add(-30 * 24 * time.Hour)
	// Legacy rows predate the ranking hook/columns.
	must(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&topics.Entity{Id: 1, Status: 1, CreatedAt: old}).Error)
	must(t, saveAndHandleRank(db, &topics.Entity{Id: 2, Status: 0, CreatedAt: old}))
	must(t, Backfill(context.Background()))
	must(t, Backfill(context.Background()))
	legacy := readTopic(t, db, 1)
	if legacy.RankScore != 0 || readSchedule(t, 1).NextRunAt != nil || legacy.PublishedAt == nil || !legacy.PublishedAt.Equal(old) {
		t.Fatalf("legacy topic made fresh: %+v", legacy)
	}
	draft := readTopic(t, db, 2)
	if draft.PublishedAt != nil {
		t.Fatal("draft has publication timestamp")
	}
	draft.Status = 1
	must(t, saveAndHandleRank(db, &draft))
	first := readTopic(t, db, 2)
	if first.PublishedAt == nil || first.PublishedAt.Before(time.Now().Add(-time.Minute)) {
		t.Fatal("first publication reused draft creation time")
	}
	must(t, Recalculate(context.Background(), 2, time.Now()))
	if readTopic(t, db, 2).RankScore != 30000 {
		t.Fatal("newly published draft missing freshness")
	}
	draft.Status = 0
	must(t, saveAndHandleRank(db, &draft))
	draft.Status = 1
	must(t, saveAndHandleRank(db, &draft))
	if !readTopic(t, db, 2).PublishedAt.Equal(*first.PublishedAt) {
		t.Fatal("republishing reset freshness")
	}
}

func TestRebuildBatchesPreservesPublicationAndCanRepeat(t *testing.T) {
	db := rankDB(t)
	published := time.Now().Add(-30 * 24 * time.Hour)
	for id := uint64(1); id <= 205; id++ {
		must(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&topics.Entity{Id: id, Status: 1, CreatedAt: published}).Error)
	}
	// SQLite can retain old decimal values when a column's affinity changes.
	// Rebuild must not try to scan those old scores into an int64.
	must(t, db.Table("topics").Where("id > 0").Updates(map[string]any{"rank_score": 12.345, "published_at": published}).Error)
	for attempt := 0; attempt < 2; attempt++ {
		var batches []int64
		count, err := Rebuild(context.Background(), func(n int64) { batches = append(batches, n) })
		must(t, err)
		if count != 205 || len(batches) != 2 || batches[0] != 200 || batches[1] != 205 {
			t.Fatalf("count %d batches %v", count, batches)
		}
		topic := readTopic(t, db, 205)
		if topic.RankScore != 0 || !topic.PublishedAt.Equal(published) || readSchedule(t, 205).NextRunAt != nil {
			t.Fatalf("unexpected rebuilt topic: %+v", topic)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Rebuild(ctx, nil); err == nil {
		t.Fatal("canceled rebuild succeeded")
	}
}

func TestRankWriteDoesNotTouchContentOrScheduleAgain(t *testing.T) {
	db := rankDB(t)
	topic := topics.Entity{Id: 1, Status: 1, Title: "unchanged"}
	must(t, saveAndHandleRank(db, &topic))
	before := readTopic(t, db, 1)
	must(t, topics.SaveRankScore(context.Background(), 1, 12345))
	after := readTopic(t, db, 1)
	if after.RankScore != 12345 || after.Title != before.Title || !after.UpdatedAt.Equal(before.UpdatedAt) || !after.PublishedAt.Equal(*before.PublishedAt) {
		t.Fatalf("rank write changed content or scheduled again: %+v", after)
	}
}

// The fixture has only topics and its scheduling table. Any dependency on reply/action/stat
// aggregation will fail, even if it is hidden behind a repository helper.
func TestRankingReadsOnlyPersistedTopicCounters(t *testing.T) {
	db := rankDB(t)
	now := time.Now()
	old := now.Add(-30 * 24 * time.Hour)
	last := now.Add(-30 * time.Minute)
	must(t, db.Create(&topics.Entity{Id: 1, Status: 1, LikeCount: 8, ReplyCount: 12, CreatedAt: old, LastPostedAt: &last}).Error)
	must(t, Recalculate(context.Background(), 1, now))
	got := readTopic(t, db, 1)
	want, _ := Score(Signals{Likes: 8, Replies: 12, LastPostedAt: &last}, old, now)
	if got.RankScore != want || readSchedule(t, 1).NextRunAt == nil {
		t.Fatalf("got %+v want %d", got, want)
	}
	must(t, Recalculate(context.Background(), 1, now.Add(24*time.Hour)))
	settled := readTopic(t, db, 1)
	if readSchedule(t, 1).NextRunAt != nil || settled.RankScore != want-20000 {
		t.Fatalf("did not settle: %+v", settled)
	}
}

func TestActivityBoundariesAndNoReplyGuard(t *testing.T) {
	now := time.Now()
	old := now.Add(-30 * 24 * time.Hour)
	baseline, _ := Score(Signals{Replies: 1}, old, now)
	for _, tc := range []struct {
		age   time.Duration
		bonus int64
	}{{0, 20}, {time.Hour, 12}, {6 * time.Hour, 6}, {12 * time.Hour, 2}, {24 * time.Hour, 0}} {
		last := now.Add(-tc.age)
		got, next := Score(Signals{Replies: 1, LastPostedAt: &last}, old, now)
		if got != baseline+tc.bonus*1000 {
			t.Fatalf("age %v score %d", tc.age, got)
		}
		if next != nil && !next.After(now) {
			t.Fatal("non-future boundary")
		}
	}
	got, next := Score(Signals{LastPostedAt: &now}, old, now)
	if got != 0 || next != nil {
		t.Fatal("first post received reply activity bonus")
	}
}

// Simulate the asynchronous consumer explicitly for isolated database tests.
func saveAndHandleRank(db *gorm.DB, topic *topics.Entity) error {
	if err := topics.SaveWithDB(db, topic); err != nil {
		return err
	}
	return topicrank.MarkAt(context.Background(), topic.Id, time.Now())
}
