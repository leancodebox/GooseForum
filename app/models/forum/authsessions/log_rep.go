package authsessions

import (
	"cmp"
	"slices"
	"time"

	"gorm.io/gorm"
)

type LogFilter struct {
	Cursor   uint64 `json:"cursor" form:"cursor"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	UserID   uint64 `json:"userId" form:"userId"`
	Result   string `json:"result" form:"result"`
	Method   string `json:"method" form:"method"`
	Since    string `json:"since" form:"since"`
	Until    string `json:"until" form:"until"`
}
type LogRow struct {
	Id            uint64    `json:"id"`
	UserId        uint64    `json:"userId"`
	Action        string    `json:"action"`
	AuthMethod    string    `json:"authMethod"`
	OAuthProvider string    `json:"oauthProvider"`
	Result        string    `json:"result"`
	Reason        string    `json:"reason,omitempty"`
	ClientIP      string    `json:"clientIp"`
	UserAgent     string    `json:"userAgent"`
	CreatedAt     time.Time `json:"createdAt"`
}
type LogPage struct {
	List       []LogRow `json:"list"`
	NextCursor uint64   `json:"nextCursor"`
	HasMore    bool     `json:"hasMore"`
	PageSize   int      `json:"pageSize"`
}

func ListLogs(filter LogFilter, owner uint64, admin bool, since, until *time.Time) (LogPage, error) {
	return listLogs(logBuilder(), filter, owner, admin, since, until)
}

func listLogs(db *gorm.DB, filter LogFilter, owner uint64, admin bool, since, until *time.Time) (LogPage, error) {
	query := db.Model(&Log{})
	if !admin {
		query = query.Where("user_id = ? AND user_id <> 0", owner)
	} else if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Method != "" {
		query = query.Where("auth_method = ?", filter.Method)
	}
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	if until != nil {
		query = query.Where("created_at <= ?", *until)
	}
	if filter.Cursor > 0 {
		query = query.Where("id < ?", filter.Cursor)
	}
	result := LogPage{List: make([]LogRow, 0), PageSize: filter.PageSize}
	query = query.Order("id DESC").Limit(filter.PageSize + 1)
	if filter.Result == "success" {
		// Keep both legacy and explicit successes bounded; an OR would sort all matching rows on some databases.
		for _, value := range []string{"success", ""} {
			var rows []LogRow
			if err := query.Session(&gorm.Session{}).Where("result = ?", value).Find(&rows).Error; err != nil {
				return LogPage{}, err
			}
			result.List = append(result.List, rows...)
		}
		slices.SortFunc(result.List, func(left, right LogRow) int { return cmp.Compare(right.Id, left.Id) })
	} else {
		if filter.Result != "" {
			query = query.Where("result = ?", filter.Result)
		}
		if err := query.Find(&result.List).Error; err != nil {
			return LogPage{}, err
		}
	}
	if len(result.List) > filter.PageSize {
		result.List = result.List[:filter.PageSize]
		result.HasMore = true
		result.NextCursor = result.List[len(result.List)-1].Id
	}
	return result, nil
}
