package accountrestrictions

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"strings"
	"testing"
)

func TestHistoryPaginationUsesUserIndexAndKeepsAccountsSeparate(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&History{}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []History{{Id: 1, UserId: 7, Reason: "first"}, {Id: 2, UserId: 8, Reason: "another account"}, {Id: 3, UserId: 7, Reason: "latest"}} {
		if err := CreateHistory(&entry); err != nil {
			t.Fatal(err)
		}
	}
	entries, total, err := HistoryPage(7, 1, 1)
	if err != nil || total != 2 || len(entries) != 1 || entries[0].Id != 3 {
		t.Fatalf("page1=%+v total=%d err=%v", entries, total, err)
	}
	entries, total, err = HistoryPage(7, 2, 1)
	if err != nil || total != 2 || len(entries) != 1 || entries[0].Id != 1 {
		t.Fatalf("page2=%+v total=%d err=%v", entries, total, err)
	}
	var plan []struct{ Detail string }
	if err = db.Raw("EXPLAIN QUERY PLAN SELECT * FROM user_restriction_history WHERE user_id = ? ORDER BY id DESC LIMIT 10", 7).Find(&plan).Error; err != nil {
		t.Fatal(err)
	}
	indexed := false
	for _, step := range plan {
		if strings.Contains(step.Detail, "idx_restriction_history_page") {
			indexed = true
		}
		if strings.Contains(step.Detail, "TEMP B-TREE") {
			t.Fatalf("history pagination requires extra sorting: %+v", plan)
		}
	}
	if !indexed {
		t.Fatalf("history pagination missed composite index: %+v", plan)
	}
}
