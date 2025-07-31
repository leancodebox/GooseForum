package vo

type SiteStats struct {
	UserCount         int64 `json:"userCount"`
	UserMonthCount    int64 `json:"userMonthCount"`
	ArticleCount      int64 `json:"articleCount"`
	ArticleMonthCount int64 `json:"articleMonthCount"`
	Reply             int64 `json:"reply"`
	LinksCount        int   `json:"linksCount"`
}
