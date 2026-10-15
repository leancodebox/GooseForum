package posts

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestReplyQuotaUsesUserDateIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Detail string }
	if err := db.Raw("EXPLAIN QUERY PLAN SELECT count(*) FROM posts WHERE user_id = ? AND post_no > 1 AND created_at > ?", 1, "2026-10-16").Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if strings.Contains(r.Detail, "idx_posts_user_created") {
			return
		}
	}
	t.Fatalf("quota index not used: %+v", rows)
}
