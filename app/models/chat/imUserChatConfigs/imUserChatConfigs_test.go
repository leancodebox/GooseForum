package imUserChatConfigs

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func TestGetUserConfigsUsesIDAsStableTieBreaker(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	const userID = uint64(981001)
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	configs := []Entity{
		{UserId: userID, PeerId: 981002, ConvId: 981002, UpdatedAt: now},
		{UserId: userID, PeerId: 981003, ConvId: 981003, UpdatedAt: now},
	}
	if err := db.Create(&configs).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("user_id = ?", userID).Delete(&Entity{}) })

	got := GetUserConfigs(userID)
	if len(got) != 2 || got[0].Id != configs[1].Id || got[1].Id != configs[0].Id {
		t.Fatalf("configs=%v, want descending id order for equal updated_at", got)
	}
}
