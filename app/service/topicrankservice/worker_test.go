package topicrankservice

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type rankSQLRecorder struct {
	logger.Interface
	clock         func() time.Time
	reads, writes int
	starts        []time.Time
	badSQL        []string
}

func (r *rankSQLRecorder) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()
	upper := strings.ToUpper(sql)
	if strings.Contains(upper, "COUNT(") || strings.Contains(upper, "SUM(") || strings.Contains(upper, "JOIN ") || strings.Contains(upper, "GROUP BY") {
		r.badSQL = append(r.badSQL, sql)
	}
	if strings.HasPrefix(upper, "SELECT") {
		r.reads++
	}
	if strings.HasPrefix(upper, "UPDATE") {
		r.writes++
		if strings.HasPrefix(upper, "UPDATE `TOPICS`") {
			r.starts = append(r.starts, r.clock())
		}
	}
}

func TestWorkerDrainsBeyondBatchLimitWithSmoothSQLBudget(t *testing.T) {
	db := rankDB(t)
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	initial := now
	rows := make([]topics.Entity, 401)
	for i := range rows {
		rows[i] = topics.Entity{Id: uint64(i + 1), Status: 0}
	}
	must(t, db.CreateInBatches(&rows, 100).Error)
	due := now.Add(-time.Hour)
	var scheduled []topicrank.Entity
	for _, row := range rows {
		scheduled = append(scheduled, topicrank.Entity{TopicID: row.Id, NextRunAt: &due, Version: 1})
	}
	must(t, topicrank.Import(context.Background(), scheduled))
	recorder := &rankSQLRecorder{Interface: db.Logger, clock: func() time.Time { return now }}
	worker := newRankWorker(20)
	worker.now = recorder.clock
	waits := 0
	worker.wait = func(_ context.Context, delay time.Duration, wake <-chan struct{}) bool {
		if wake != nil {
			if recorder.writes != 2*len(rows) {
				t.Fatalf("idle with backlog: processed %d", recorder.writes)
			}
			return false
		}
		if delay != 50*time.Millisecond {
			t.Fatalf("unexpected throttle/backoff: %v", delay)
		}
		waits++
		now = now.Add(delay)
		return true
	}
	originalLogger := db.Logger
	db.Logger = recorder
	t.Cleanup(func() { db.Logger = originalLogger })
	worker.run(context.Background(), make(chan struct{}))
	if recorder.writes != 802 || waits != 400 || now.Sub(initial) != 20*time.Second {
		t.Fatalf("writes=%d waits=%d elapsed=%v", recorder.writes, waits, now.Sub(initial))
	}
	// 401 primary-key reads, 21 schedule batches and one final empty read.
	if recorder.reads != 423 || len(recorder.badSQL) != 0 {
		t.Fatalf("reads=%d forbidden SQL=%v", recorder.reads, recorder.badSQL)
	}
	for i := 1; i < len(recorder.starts); i++ {
		if recorder.starts[i].Sub(recorder.starts[i-1]) < 50*time.Millisecond {
			t.Fatal("worker emitted a burst")
		}
	}
}

func TestWorkerWaitsForDeadlineAndWakeupCannotBypassRate(t *testing.T) {
	db := rankDB(t)
	now := time.Now().Truncate(time.Millisecond)
	future := now.Add(2 * time.Second)
	must(t, db.Create(&topics.Entity{Id: 1}).Error)
	must(t, topicrank.MarkAt(context.Background(), 1, future))
	worker := newRankWorker(20)
	worker.now = func() time.Time { return now }
	wake := make(chan struct{}, 1)
	phases := 0
	worker.wait = func(_ context.Context, delay time.Duration, ch <-chan struct{}) bool {
		phases++
		switch phases {
		case 1:
			if ch != wake || delay != 2*time.Second {
				t.Fatalf("future wait=%v", delay)
			}
			// New activity advances the deadline and wakes an idle consumer.
			must(t, topicrank.MarkAt(context.Background(), 1, now))
			return true
		case 2:
			if ch != wake || delay != rankIdleWait {
				t.Fatal("expected empty queue wait")
			}
			must(t, topicrank.MarkAt(context.Background(), 1, now))
			return true
		case 3:
			if ch != nil || delay != 50*time.Millisecond {
				t.Fatal("wakeup bypassed pacing")
			}
			now = now.Add(delay)
			return true
		default:
			return false
		}
	}
	worker.run(context.Background(), wake)
	if phases != 4 {
		t.Fatalf("wait phases=%d", phases)
	}
}

func TestWorkerDefersPoisonedRowsAndContinues(t *testing.T) {
	db := rankDB(t)
	now := time.Now().Truncate(time.Millisecond)
	for id := uint64(1); id <= 3; id++ {
		must(t, db.Create(&topics.Entity{Id: id}).Error)
	}
	for id := uint64(1); id <= 3; id++ {
		must(t, topicrank.MarkAt(context.Background(), id, now.Add(-time.Hour)))
	}
	must(t, db.Table("topics").Where(queryopt.Eq("id", 1)).UpdateColumn("like_count", "invalid").Error)
	worker := newRankWorker(20)
	worker.now = func() time.Time { return now }
	worker.wait = func(_ context.Context, d time.Duration, ch <-chan struct{}) bool {
		if ch != nil {
			return false
		}
		now = now.Add(d)
		return true
	}
	worker.run(context.Background(), make(chan struct{}))
	pending, err := topicrank.Pending(context.Background(), 10)
	must(t, err)
	if len(pending) != 1 || pending[0].ID != 1 || !pending[0].At.After(now) {
		t.Fatalf("failed row blocked/lost: %v", pending)
	}
	// The schedule is persisted; a restarted worker can retry after repair.
	must(t, db.Table("topics").Where(queryopt.Eq("id", 1)).UpdateColumn("like_count", 0).Error)
	now = pending[0].At
	worker.run(context.Background(), make(chan struct{}))
	pending, err = topicrank.Pending(context.Background(), 10)
	must(t, err)
	if len(pending) != 0 {
		t.Fatalf("restart left work: %v", pending)
	}
}

func TestWorkerShutdownAndDatabaseBackoff(t *testing.T) {
	db := rankDB(t)
	must(t, db.Migrator().DropTable(&topicrank.Entity{}))
	worker := newRankWorker(20)
	calls := 0
	worker.wait = func(_ context.Context, d time.Duration, wake <-chan struct{}) bool {
		calls++
		if d != rankDatabaseBackoff || wake != nil {
			t.Fatal("database retry can spin or wake early")
		}
		return false
	}
	worker.run(context.Background(), make(chan struct{}))
	if calls != 1 {
		t.Fatalf("backoff calls=%d", calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if waitRankWorker(ctx, time.Hour, nil) {
		t.Fatal("shutdown ignored")
	}
	worker.run(ctx, nil)
	if calls != 1 {
		t.Fatal("canceled worker continued")
	}
}

func TestMarkPreservesQueueAgeAndFailedRetryDoesNotOverwriteNewSchedule(t *testing.T) {
	db := rankDB(t)
	now := time.Now().Truncate(time.Millisecond)
	must(t, db.Create(&topics.Entity{Id: 1, Status: 1}).Error)
	must(t, topicrank.MarkAt(context.Background(), 1, now))
	must(t, topicrank.MarkAt(context.Background(), 1, now.Add(time.Minute)))
	entries, err := topicrank.Pending(context.Background(), 1)
	must(t, err)
	if len(entries) != 1 || !entries[0].At.Equal(now) {
		t.Fatal("activity moved a waiting topic backwards")
	}
	earlier := now.Add(-time.Minute)
	must(t, topicrank.MarkAt(context.Background(), 1, earlier))
	must(t, topicrank.Defer(context.Background(), entries[0], now.Add(time.Hour)))
	entries, err = topicrank.Pending(context.Background(), 1)
	must(t, err)
	if !entries[0].At.Equal(earlier) {
		t.Fatal("failed retry overwrote a changed schedule")
	}
}

func TestNewTriggerSurvivesCompletionAndSettledVersionDoesNotReset(t *testing.T) {
	db := rankDB(t)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)
	must(t, db.Create(&topics.Entity{Id: 1, Status: 1, CreatedAt: now.Add(-30 * 24 * time.Hour)}).Error)
	must(t, topicrank.MarkAt(ctx, 1, now))
	entries, err := topicrank.Pending(ctx, 1)
	must(t, err)
	original := entries[0]
	// Interleave a producer after the score UPDATE, before task acknowledgement.
	triggered := false
	must(t, db.Callback().Update().After("gorm:commit_or_rollback_transaction").Register("test:new_rank_trigger", func(tx *gorm.DB) {
		if tx.Statement.Table == "topics" && !triggered {
			triggered = true
			must(t, topicrank.MarkAt(ctx, 1, now))
		}
	}))
	t.Cleanup(func() { db.Callback().Update().Remove("test:new_rank_trigger") })
	must(t, recalculateScheduled(ctx, original, now))
	current := readSchedule(t, 1)
	if !triggered || current.Version <= original.Version || current.NextRunAt == nil {
		t.Fatal("worker acknowledged a newer trigger")
	}
	must(t, db.Callback().Update().Remove("test:new_rank_trigger"))
	must(t, recalculateScheduled(ctx, topicrank.ScheduledTopic{ID: 1, Version: current.Version}, now))
	settled := readSchedule(t, 1)
	if settled.NextRunAt != nil || settled.Version <= current.Version {
		t.Fatal("settling removed version history")
	}
	must(t, topicrank.MarkAt(ctx, 1, now))
	must(t, topicrank.SetNext(ctx, original, nil))
	fresh := readSchedule(t, 1)
	if fresh.Version <= settled.Version || fresh.NextRunAt == nil {
		t.Fatal("old completion erased a reactivated task")
	}
}

func TestTopicWriteFailureCanDeferIndependentlyOfOtherTopics(t *testing.T) {
	db := rankDB(t)
	now := time.Now().Truncate(time.Millisecond)
	for id := uint64(1); id <= 2; id++ {
		must(t, db.Create(&topics.Entity{Id: id}).Error)
		must(t, topicrank.MarkAt(context.Background(), id, now.Add(-time.Hour)))
	}
	must(t, db.Exec("CREATE TRIGGER reject_topic_rank BEFORE UPDATE ON topics WHEN OLD.id=1 BEGIN SELECT RAISE(FAIL, 'one topic rejects writes'); END").Error)
	worker := newRankWorker(20)
	worker.now = func() time.Time { return now }
	worker.wait = func(_ context.Context, d time.Duration, wake <-chan struct{}) bool {
		if wake != nil {
			return false
		}
		now = now.Add(d)
		return true
	}
	worker.run(context.Background(), make(chan struct{}))
	bad, good := readSchedule(t, 1), readSchedule(t, 2)
	if bad.NextRunAt == nil || !bad.NextRunAt.After(now) || good.NextRunAt != nil {
		t.Fatalf("bad topic blocked healthy work: bad=%+v good=%+v", bad, good)
	}
}
