package badgeservice

import "testing"

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
