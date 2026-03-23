package vo

type SiteStats struct {
	UserCount         uint64 `json:"userCount"`
	UserMonthCount    int64  `json:"userMonthCount"`
	ArticleCount      uint64 `json:"articleCount"`
	ArticleMonthCount int64  `json:"articleMonthCount"`
	Reply             uint64 `json:"reply"`
	LinksCount        int    `json:"linksCount"`
}
