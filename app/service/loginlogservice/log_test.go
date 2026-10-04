package loginlogservice

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
)

func TestFailureBudgetBoundsAndExpiry(t *testing.T) {
	b := failureBudget{ips: make(map[string]window)}
	now := time.Now()
	for i := 0; i < 10; i++ {
		if !b.allow("one", now) {
			t.Fatal("premature rejection")
		}
	}
	if b.allow("one", now) {
		t.Fatal("IP exceeded its budget")
	}
	for i := 0; i < 990; i++ {
		if !b.allow(fmt.Sprint(i), now) {
			t.Fatal("premature global rejection")
		}
	}
	if b.allow("another", now) {
		t.Fatal("global exceeded its budget")
	}
	if !b.allow("one", now.Add(time.Minute)) {
		t.Fatal("budget did not reset")
	}
	if len(b.ips) != 1 {
		t.Fatal("expired windows not removed")
	}
}

func TestConcurrentFailureBudget(t *testing.T) {
	b := failureBudget{ips: make(map[string]window)}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); b.allow("ip", time.Now()) }()
	}
	wg.Wait()
	if b.global.count != 10 {
		t.Fatalf("got %d accepted failures", b.global.count)
	}
}

func TestAccountBudgetAcrossIPs(t *testing.T) {
	b := failureBudget{ips: make(map[string]window)}
	now := time.Now()
	for i := 0; i < 10; i++ {
		if !b.allow(fmt.Sprint(i), now, "digest") {
			t.Fatal("premature account rejection")
		}
	}
	if b.allow("new-ip", now, "digest") {
		t.Fatal("account exceeded budget across IPs")
	}
	if !b.allow("new-ip", now, "other-digest") {
		t.Fatal("other account rejected")
	}
}

func TestHistoryOwnerIsolationAndFilters(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.Migrator().DropTable(&authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	rows := []authsessions.Log{
		{UserId: 7, Action: "login_success", AuthMethod: "password", Result: "success", Reason: "internal", CreatedAt: now},
		{UserId: 8, Action: "login_success", AuthMethod: "oauth", Result: "success", CreatedAt: now},
		{UserId: 0, Action: "login_failure", AuthMethod: "password", Result: "failure", Reason: "credentials_rejected", CreatedAt: now},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	own, err := List(Filter{UserID: 8}, 7, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(own.List) != 1 || own.HasMore || own.NextCursor != 0 || own.PageSize != 20 || own.List[0].UserId != 7 || own.List[0].Reason != "" {
		t.Fatalf("owner leaked foreign records/reason: %+v", own)
	}
	unknown, err := List(Filter{}, 0, false)
	if err != nil || len(unknown.List) != 0 || unknown.HasMore || unknown.NextCursor != 0 {
		t.Fatalf("anonymous user got audit events: %+v %v", unknown, err)
	}
	failures, err := List(Filter{Result: "failure", Method: "password", Since: now.Add(-time.Hour).Format(time.RFC3339), PageSize: 999}, 0, true)
	if err != nil || len(failures.List) != 1 || failures.HasMore || failures.NextCursor != 0 || failures.PageSize != 100 || failures.List[0].Reason != "credentials_rejected" {
		t.Fatalf("admin filters invalid: %+v %v", failures, err)
	}
	if _, err := List(Filter{Since: "bad-date"}, 0, true); err == nil {
		t.Fatal("invalid date accepted")
	}
}

func TestCursorPagesKeepOwnerIsolationAndHideInternalReasons(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.Migrator().DropTable(&authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = connection.Close() })
	if err := db.AutoMigrate(&authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	for _, owner := range []uint64{7, 8, 7, 8, 7, 8} {
		if err := db.Create(&authsessions.Log{UserId: owner, AuthMethod: "password", Result: "failure", Reason: "internal", CreatedAt: time.Now()}).Error; err != nil {
			t.Fatal(err)
		}
	}
	filter := Filter{UserID: 8, PageSize: 1, Result: "failure", Method: "password"}
	for _, expectedID := range []uint64{5, 3, 1} {
		page, err := List(filter, 7, false)
		if err != nil || len(page.List) != 1 || page.List[0].Id != expectedID || page.List[0].UserId != 7 || page.List[0].Reason != "" || page.HasMore != (expectedID > 1) {
			t.Fatalf("owner cursor page expected=%d page=%+v err=%v", expectedID, page, err)
		}
		if expectedID > 1 {
			if page.NextCursor != expectedID {
				t.Fatalf("wrong continuation: %+v", page)
			}
			filter.Cursor = page.NextCursor
		} else if page.NextCursor != 0 {
			t.Fatalf("finished page has continuation: %+v", page)
		}
	}
}

func TestClipPreservesUTF8(t *testing.T) {
	if got := clip(strings.Repeat("文", 200), 512); len(got) > 512 || strings.ContainsRune(got, '\uFFFD') {
		t.Fatal("invalid truncated UTF8")
	}
}
