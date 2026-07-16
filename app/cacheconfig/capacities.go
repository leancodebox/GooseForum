// Package cacheconfig owns the capacity policy for bounded server-side caches.
package cacheconfig

// Capacities contains entry limits, not memory sizes.
type Capacities struct {
	DefaultLocal       uint64
	TopicList          uint64
	Category           uint64
	SiteStatistics     uint64
	PageConfig         uint64
	UserInfo           uint64
	UserPublicProfile  uint64
	UserActivity       uint64
	RolePermission     uint64
	UnreadStatus       uint64
	ModerationStatus   uint64
	ConversationAccess uint64
	BadgeDefinitions   uint64
	RuntimeTheme       uint64
	SEOXML             uint64
}

var current = Capacities{
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

// Current returns the process cache capacity policy.
func Current() Capacities {
	return current
}
