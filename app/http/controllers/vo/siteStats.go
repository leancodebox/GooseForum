package vo

// SiteStats contains site-wide counters shown in shared page components.
type SiteStats struct {
	UserCount       uint64 `json:"userCount"`
	UserMonthCount  int64  `json:"userMonthCount"`
	TopicMaxID      uint64 `json:"topicMaxId"`
	TopicMonthCount int64  `json:"topicMonthCount"`
	PostMaxID       uint64 `json:"postMaxId"`
	LinksCount      int    `json:"linksCount"`
}
