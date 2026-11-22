package userservice

import (
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func TestUserInfoLookupCachesOnlyExistingUsers(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: "recipient-cache-test", Email: "recipient-cache@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		InvalidateUserCaches(user.Id)
		db.Unscoped().Delete(&user)
	})
	if _, ok := GetUserInfo(user.Id); !ok {
		t.Fatal("failed to prime the shared user cache")
	}
	if err := db.Unscoped().Delete(&user).Error; err != nil {
		t.Fatal(err)
	}
	if info, err := GetUserInfoWithError(user.Id); err != nil || info.Id != user.Id {
		t.Fatalf("shared cache lookup returned id %d, error %v", info.Id, err)
	}
	InvalidateUserCaches(user.Id)
	if _, err := GetUserInfoWithError(user.Id); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected missing user after invalidation, got %v", err)
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := GetUserInfoWithError(user.Id); err != nil {
		t.Fatalf("previous miss hid newly created user: %v", err)
	}
}

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
