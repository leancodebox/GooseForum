package cacheconfig

import "testing"

func TestCurrentCapacitiesPreserveExistingLimits(t *testing.T) {
	got := Current()
	want := Capacities{
		DefaultLocal:       2048,
		TopicList:          512,
		Category:           4,
		SiteStatistics:     4,
		PageConfig:         4,
		UserInfo:           2048,
		UserPublicProfile:  1024,
		UserActivity:       8192,
		RolePermission:     256,
		UnreadStatus:       2048,
		ModerationStatus:   2048,
		ConversationAccess: 4096,
		BadgeDefinitions:   4,
		RuntimeTheme:       1,
		SEOXML:             128,
	}
	if got != want {
		t.Fatalf("Current() = %+v, want %+v", got, want)
	}
}
