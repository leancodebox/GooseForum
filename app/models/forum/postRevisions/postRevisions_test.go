package postRevisions_test

import (
	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"gorm.io/gorm"
	"testing"
)

func TestRetentionKeepsNewestVersionsAcrossPosts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&postRevisions.Entity{}); err != nil {
		t.Fatal(err)
	}
	// Insert in reverse completion order to model delayed snapshot writes.
	for post := uint64(1); post <= 203; post++ {
		for version := uint64(25); version > 0; version-- {
			if err = db.Create(&postRevisions.Entity{PostId: post, RevisionNo: version}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	for run := 0; run < 2; run++ {
		if err = postRevisions.PruneAllWithDB(db); err != nil {
			t.Fatal(err)
		}
	}
	for post := uint64(1); post <= 203; post++ {
		page, err := postRevisions.ListPageWithDB(db, postRevisions.PageQuery{PostId: post})
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) != 10 || page.Items[0].RevisionNo != 25 || page.Items[9].RevisionNo != 16 {
			t.Fatalf("post %d: %+v", post, page)
		}
	}
}
