package postRevisions_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"gorm.io/gorm"
)

func TestListPageWithDBUsesDescendingRevisionCursor(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&postRevisions.Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	for revision := uint64(1); revision <= 3; revision++ {
		if err := db.Create(&postRevisions.Entity{PostId: 7, RevisionNo: revision, Content: "v"}).Error; err != nil {
			t.Fatalf("create revision %d: %v", revision, err)
		}
	}
	page, err := postRevisions.ListPageWithDB(db, postRevisions.PageQuery{PostId: 7, PageSize: 2})
	if err != nil || len(page.Items) != 2 || page.Items[0].RevisionNo != 3 || page.Items[1].RevisionNo != 2 || !page.HasMore || page.NextCursor != 2 {
		t.Fatalf("first page = %+v, err=%v", page, err)
	}
	page, err = postRevisions.ListPageWithDB(db, postRevisions.PageQuery{PostId: 7, Cursor: page.NextCursor, PageSize: 2})
	if err != nil || len(page.Items) != 1 || page.Items[0].RevisionNo != 1 || page.HasMore {
		t.Fatalf("second page = %+v, err=%v", page, err)
	}
	payload, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("marshal page: %v", err)
	}
	if string(payload) == "" || string(payload[:1]) != "{" || !containsJSONKey(payload, `"nextCursor"`) {
		t.Fatalf("page JSON = %s, want lower camel case keys", payload)
	}
}

func containsJSONKey(payload []byte, key string) bool {
	return bytes.Contains(payload, []byte(key))
}
