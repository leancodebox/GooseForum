package topicunseenservice

import (
	"os"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/service/kvstore"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "topic-unseen-test-")
	if err != nil {
		panic(err)
	}
	preferences.Set("badger.path", dir)
	code := m.Run()
	kvstore.Close()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func TestNextTrackingKeepsLastSeenDuringActiveVisit(t *testing.T) {
	start := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	current := trackingState{LastActiveAt: start, LastSeenAt: start.Add(-time.Hour)}

	next := nextTracking(current, start.Add(10*time.Minute))
	if !next.LastActiveAt.Equal(start.Add(10 * time.Minute)) {
		t.Fatalf("LastActiveAt = %v", next.LastActiveAt)
	}
	if !next.LastSeenAt.Equal(current.LastSeenAt) {
		t.Fatalf("LastSeenAt = %v, want %v", next.LastSeenAt, current.LastSeenAt)
	}
}

func TestBadgerTrackingFlow(t *testing.T) {
	start := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	activity := TopicActivity{TopicID: 42, LastPostedAt: start.Add(-time.Hour)}

	first, err := Resolve(7, []TopicActivity{activity}, start)
	if err != nil {
		t.Fatalf("first Resolve() error = %v", err)
	}
	if !first[42] {
		t.Fatal("first Resolve() unseen = false, want true")
	}

	refreshed, err := Resolve(7, []TopicActivity{activity}, start.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("refreshed Resolve() error = %v", err)
	}
	if !refreshed[42] {
		t.Fatal("refreshed Resolve() cleared unseen during active visit")
	}

	if err := MarkVisited(7, 42, start.Add(6*time.Minute)); err != nil {
		t.Fatalf("MarkVisited() error = %v", err)
	}
	visited, err := Resolve(7, []TopicActivity{activity}, start.Add(7*time.Minute))
	if err != nil {
		t.Fatalf("visited Resolve() error = %v", err)
	}
	if visited[42] {
		t.Fatal("visited Resolve() unseen = true, want false")
	}

	activity.LastPostedAt = start.Add(8 * time.Minute)
	updated, err := Resolve(7, []TopicActivity{activity}, start.Add(9*time.Minute))
	if err != nil {
		t.Fatalf("updated Resolve() error = %v", err)
	}
	if !updated[42] {
		t.Fatal("updated Resolve() unseen = false after a new post")
	}
}

func TestMarkVisitedDoesNotMoveBackward(t *testing.T) {
	latest := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	if err := MarkVisited(7, 42, latest); err != nil {
		t.Fatalf("latest MarkVisited() error = %v", err)
	}
	if err := MarkVisited(7, 42, latest.Add(-time.Minute)); err != nil {
		t.Fatalf("older MarkVisited() error = %v", err)
	}

	value, err := kvstore.GetBytes(visitKey(7, 42))
	if err != nil {
		t.Fatalf("GetBytes() error = %v", err)
	}
	visitedAt, ok := decodeVisit(value)
	if !ok || !visitedAt.Equal(latest) {
		t.Fatalf("visitedAt = %v, want %v", visitedAt, latest)
	}
}

func TestNextTrackingMovesLastSeenAfterInactiveGap(t *testing.T) {
	start := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	current := trackingState{LastActiveAt: start, LastSeenAt: start.Add(-time.Hour)}

	next := nextTracking(current, start.Add(31*time.Minute))
	if !next.LastSeenAt.Equal(start) {
		t.Fatalf("LastSeenAt = %v, want %v", next.LastSeenAt, start)
	}
}

func TestMinLastSeenAtUsesTwoDayFloor(t *testing.T) {
	now := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	if got := minLastSeenAt(time.Time{}, now); !got.Equal(now.Add(-trackingTTL)) {
		t.Fatalf("minLastSeenAt(zero) = %v", got)
	}
	recent := now.Add(-time.Hour)
	if got := minLastSeenAt(recent, now); !got.Equal(recent) {
		t.Fatalf("minLastSeenAt(recent) = %v", got)
	}
}

func TestResolveUnseenUsesLastPostedAndLastVisited(t *testing.T) {
	now := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	baseline := now.Add(-12 * time.Hour)
	activities := []TopicActivity{
		{TopicID: 1, LastPostedAt: now.Add(-time.Hour)},
		{TopicID: 2, LastPostedAt: now.Add(-time.Hour)},
		{TopicID: 3, LastPostedAt: now.Add(-24 * time.Hour)},
	}
	visited := map[uint64]time.Time{
		2: now.Add(-30 * time.Minute),
	}

	got := resolveUnseen(activities, visited, baseline)
	if !got[1] {
		t.Fatal("topic 1 unseen = false, want true")
	}
	if got[2] {
		t.Fatal("topic 2 unseen = true after newer visit")
	}
	if got[3] {
		t.Fatal("topic 3 unseen = true before baseline")
	}
}
