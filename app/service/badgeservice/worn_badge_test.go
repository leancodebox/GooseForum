package badgeservice

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
)

func TestWornBadgeRequiresWearableBadge(t *testing.T) {
	items := []UserBadge{
		{Badge: Badge{Code: "plain", IsEnabled: true, IsWearable: false}},
		{Badge: Badge{Code: "wearable", IsEnabled: true, IsWearable: true}},
	}

	if got := WornBadgeFromList(items, "plain"); got != nil {
		t.Fatalf("WornBadgeFromList() returned non-wearable badge: %#v", got)
	}
	if got := WornBadgeFromList(items, "wearable"); got == nil || got.Code != "wearable" {
		t.Fatalf("WornBadgeFromList() = %#v, want wearable badge", got)
	}
}

func TestWornBadgesFromRecordsMatchesSelectedActiveBadge(t *testing.T) {
	selected := map[uint64]string{1: "wearable", 2: "plain", 3: "wearable"}
	records := []*userBadges.Entity{
		{UserId: 1, BadgeCode: "other"},
		{UserId: 1, BadgeCode: "wearable"},
		{UserId: 2, BadgeCode: "plain"},
	}
	definitions := map[string]Badge{
		"wearable": {Code: "wearable", IsEnabled: true, IsWearable: true},
		"plain":    {Code: "plain", IsEnabled: true, IsWearable: false},
		"other":    {Code: "other", IsEnabled: true, IsWearable: true},
	}

	got := wornBadgesFromRecords(selected, records, definitions)
	if len(got) != 1 || got[1] == nil || got[1].Code != "wearable" {
		t.Fatalf("wornBadgesFromRecords() = %#v, want only user 1 wearable badge", got)
	}
}
