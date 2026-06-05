package userservice

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
)

func TestBuildUserCardUsesRecentActivityCache(t *testing.T) {
	userID := uint64(777001)
	oldStore := activityStore
	testStore := newTestActivityStore(t)
	activityStore = testStore
	defer func() {
		activityStore = oldStore
	}()

	stored := time.Now().Add(-10 * time.Minute)
	recent := time.Now()
	rememberUserActivity(userID, recent)

	card := buildUserCard(UserPublicProfile{
		User: UserPublicInfo{Id: userID, Username: "activity-user"},
		Stats: userStatistics.Entity{
			UserId:         userID,
			LastActiveTime: stored,
		},
	})

	if !card.IsOnline {
		t.Fatal("buildUserCard() did not mark recent activity user online")
	}
	if !card.LastActiveTime.Equal(recent) {
		t.Fatalf("LastActiveTime = %v, want recent %v", card.LastActiveTime, recent)
	}
}
