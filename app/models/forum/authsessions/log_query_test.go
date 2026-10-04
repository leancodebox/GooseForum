package authsessions

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func logQueryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.AutoMigrate(&Log{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestLogHistoryIndexesCoverIDCursorAndFilters(t *testing.T) {
	db := logQueryDB(t)
	for _, query := range []struct {
		condition, index string
		args             []any
	}{
		{"user_id = ? AND id < ?", "idx_auth_log_user_id", []any{7, 100}},
		{"result = ? AND id < ?", "idx_auth_log_result_id", []any{"failure", 100}},
		{"auth_method = ? AND id < ?", "idx_auth_log_method_id", []any{"password", 100}},
		{"result = ? AND id < ? AND created_at >= ? AND created_at <= ?", "idx_auth_log_result_id", []any{"success", 100, time.Now().Add(-time.Hour), time.Now()}},
		{"result = ? AND id < ? AND created_at >= ? AND created_at <= ?", "idx_auth_log_result_id", []any{"", 100, time.Now().Add(-time.Hour), time.Now()}},
		{"id < ?", "INTEGER PRIMARY KEY", []any{100}},
	} {
		var plan []struct{ Detail string }
		sql := "EXPLAIN QUERY PLAN SELECT id, user_id, action FROM user_auth_token_logs WHERE " + query.condition + " ORDER BY id DESC LIMIT 20"
		if err := db.Raw(sql, query.args...).Scan(&plan).Error; err != nil {
			t.Fatal(err)
		}
		details := ""
		for _, row := range plan {
			details += row.Detail + "\n"
		}
		if !strings.Contains(details, query.index) || strings.Contains(details, "TEMP B-TREE") {
			t.Fatalf("inefficient history plan: %s", details)
		}
	}
}

func TestSuccessCursorMergesLegacyRowsWithOwnerMethodAndDateFilters(t *testing.T) {
	db := logQueryDB(t)
	now := time.Now().UTC()
	rows := []Log{
		{UserId: 7, Result: "", AuthMethod: "password", CreatedAt: now},
		{UserId: 7, Result: "success", AuthMethod: "password", CreatedAt: now},
		{UserId: 8, Result: "", AuthMethod: "password", CreatedAt: now},
		{UserId: 7, Result: "success", AuthMethod: "oauth", CreatedAt: now},
		{UserId: 7, Result: "", AuthMethod: "password", CreatedAt: now.Add(-2 * time.Hour)},
		{UserId: 7, Result: "failure", AuthMethod: "password", CreatedAt: now},
		{UserId: 7, Result: "success", AuthMethod: "password", CreatedAt: now},
		{UserId: 7, Result: "", AuthMethod: "password", CreatedAt: now.Add(-time.Minute)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	type capturedQuery struct {
		sql  string
		args []any
	}
	var queries []capturedQuery
	if err := db.Callback().Query().After("gorm:query").Register("test:capture-success-cursor", func(tx *gorm.DB) {
		queries = append(queries, capturedQuery{tx.Statement.SQL.String(), append([]any(nil), tx.Statement.Vars...)})
	}); err != nil {
		t.Fatal(err)
	}
	since, until := now.Add(-time.Hour), now.Add(time.Hour)
	filter := LogFilter{PageSize: 2, Result: "success", Method: "password", UserID: 8}
	first, err := listLogs(db, filter, 7, false, &since, &until)
	if err != nil || len(first.List) != 2 || first.List[0].Id != 8 || first.List[1].Id != 7 || !first.HasMore || first.NextCursor != 7 {
		t.Fatalf("merged first page=%+v err=%v", first, err)
	}
	filter.Cursor = first.NextCursor
	second, err := listLogs(db, filter, 7, false, &since, &until)
	if err != nil || len(second.List) != 2 || second.List[0].Id != 2 || second.List[1].Id != 1 || second.HasMore || second.NextCursor != 0 {
		t.Fatalf("merged last page=%+v err=%v", second, err)
	}
	if len(queries) != 4 {
		t.Fatalf("success compatibility must use two bounded queries per page: %+v", queries)
	}
	for _, query := range queries {
		sql := strings.ToUpper(query.sql)
		if strings.Contains(sql, "COUNT(") || strings.Contains(sql, "OFFSET") || !strings.Contains(sql, "ORDER BY ID DESC LIMIT 3") {
			t.Fatalf("unbounded success query: %s", query.sql)
		}
		var plan []struct{ Detail string }
		if err := db.Raw("EXPLAIN QUERY PLAN "+query.sql, query.args...).Scan(&plan).Error; err != nil {
			t.Fatal(err)
		}
		details := ""
		for _, step := range plan {
			details += step.Detail + "\n"
		}
		if strings.Contains(details, "TEMP B-TREE") || !strings.Contains(details, "idx_auth_log_") {
			t.Fatalf("success/date/owner query sorted outside index: %s", details)
		}
	}
}

func TestLogCursorPaginationIgnoresTimestampOrderAndNewerInserts(t *testing.T) {
	db := logQueryDB(t)
	now := time.Now()
	for _, at := range []time.Time{now, now, now.Add(-time.Hour), now, now.Add(-2 * time.Hour), now, now.Add(-3 * time.Hour)} {
		if err := db.Create(&Log{UserId: 7, CreatedAt: at}).Error; err != nil {
			t.Fatal(err)
		}
	}
	first, err := listLogs(db, LogFilter{PageSize: 3}, 0, true, nil, nil)
	if err != nil || !first.HasMore || first.NextCursor != 5 || first.PageSize != 3 {
		t.Fatalf("first page=%+v err=%v", first, err)
	}
	var ids []uint64
	for _, row := range first.List {
		ids = append(ids, row.Id)
	}
	if !slices.Equal(ids, []uint64{7, 6, 5}) {
		t.Fatalf("timestamp changed ID order: %v", ids)
	}
	if err := db.Create(&Log{UserId: 7, CreatedAt: now.Add(-4 * time.Hour)}).Error; err != nil {
		t.Fatal(err)
	}
	cursor := first.NextCursor
	for {
		page, err := listLogs(db, LogFilter{Cursor: cursor, PageSize: 3}, 0, true, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range page.List {
			ids = append(ids, row.Id)
		}
		if !page.HasMore {
			if page.NextCursor != 0 {
				t.Fatalf("last page retained next cursor: %+v", page)
			}
			break
		}
		if page.NextCursor >= cursor {
			t.Fatalf("cursor failed to advance: %+v", page)
		}
		cursor = page.NextCursor
	}
	if !slices.Equal(ids, []uint64{7, 6, 5, 4, 3, 2, 1}) {
		t.Fatalf("duplicate or skipped row after insert: %v", ids)
	}
	empty, err := listLogs(db, LogFilter{Cursor: 1, PageSize: 3}, 0, true, nil, nil)
	if err != nil || empty.List == nil || len(empty.List) != 0 || empty.HasMore || empty.NextCursor != 0 {
		t.Fatalf("empty page=%+v err=%v", empty, err)
	}
	latest, err := listLogs(db, LogFilter{PageSize: 3}, 0, true, nil, nil)
	if err != nil || len(latest.List) == 0 || latest.List[0].Id != 8 {
		t.Fatalf("refresh did not find new insert: %+v err=%v", latest, err)
	}
}

func TestLogCursorQueryUsesOneProjectedSelectWithoutCountOrOffset(t *testing.T) {
	db := logQueryDB(t)
	var queries []string
	if err := db.Callback().Query().After("gorm:query").Register("test:capture-log-cursor", func(tx *gorm.DB) { queries = append(queries, tx.Statement.SQL.String()) }); err != nil {
		t.Fatal(err)
	}
	if _, err := listLogs(db, LogFilter{Cursor: 500, PageSize: 20, Method: "password"}, 7, false, nil, nil); err != nil {
		t.Fatal(err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected one query, got %v", queries)
	}
	sql := strings.ToUpper(queries[0])
	if strings.Contains(sql, "COUNT(") || strings.Contains(sql, "OFFSET") || !strings.Contains(sql, "ID <") || !strings.Contains(sql, "ORDER BY ID DESC LIMIT 21") {
		t.Fatalf("unexpected cursor SQL: %s", queries[0])
	}
	if strings.Contains(sql, "SELECT *") || strings.Contains(sql, "USER_AUTH_TOKEN_ID") {
		t.Fatalf("full log model queried instead of read projection: %s", queries[0])
	}
}
