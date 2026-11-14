package posts

import (
	"errors"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"gorm.io/gorm"
)

func TestSaveReviewedWithRevisionStoresPreviousContent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Entity{}, &postRevisions.Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	post := &Entity{TopicId: 1, PostNo: 2, UserId: 7, Content: "before", SourceVersion: 1}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	post.Content, post.SourceVersion, post.ModerationVersion = "after", 2, 1
	if err := SaveReviewedWithRevisionWithDB(db, post, 0); err != nil {
		t.Fatalf("save reviewed: %v", err)
	}
	var revision postRevisions.Entity
	if err := db.Where("post_id = ?", post.Id).First(&revision).Error; err != nil {
		t.Fatalf("load revision: %v", err)
	}
	if revision.RevisionNo != 1 || revision.Content != "before" || revision.SourceVersion != 1 || revision.AuthorId != 7 {
		t.Fatalf("revision = %+v, want previous source snapshot", revision)
	}
}

func TestSaveReviewedDoesNotCreateRevisionForModerationOnlyUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Entity{}, &postRevisions.Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	post := &Entity{TopicId: 1, PostNo: 2, UserId: 7, Content: "same"}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	post.ModerationVersion = 1
	if err := SaveReviewedWithDB(db, post, 0); err != nil {
		t.Fatalf("save reviewed: %v", err)
	}
	var count int64
	if err := db.Model(&postRevisions.Entity{}).Where("post_id = ?", post.Id).Count(&count).Error; err != nil {
		t.Fatalf("count revisions: %v", err)
	}
	if count != 0 {
		t.Fatalf("revision count = %d, want 0", count)
	}
}

func TestSaveReviewedWithRevisionSucceedsWhenHistoryUnavailable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate posts: %v", err)
	}
	post := &Entity{TopicId: 1, PostNo: 2, UserId: 7, Content: "before"}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	post.Content, post.ModerationVersion = "after", 1
	if err := SaveReviewedWithRevisionWithDB(db, post, 0); err != nil {
		t.Fatalf("history failure must not fail editing: %v", err)
	}
	var stored Entity
	if err := db.First(&stored, post.Id).Error; err != nil {
		t.Fatalf("load post after history failure: %v", err)
	}
	if stored.Content != "after" || stored.ModerationVersion != 1 {
		t.Fatalf("post after history failure = %+v, want saved content and version", stored)
	}
}

func TestRevisionRetentionAndConflictWithoutTransaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&Entity{}, &postRevisions.Entity{}); err != nil {
		t.Fatal(err)
	}
	post := &Entity{TopicId: 1, PostNo: 2, UserId: 7, Content: "initial"}
	if err = db.Create(post).Error; err != nil {
		t.Fatal(err)
	}
	check := func(tx *gorm.DB) {
		if _, transactional := tx.Statement.ConnPool.(gorm.TxCommitter); transactional {
			t.Error("ordinary edit/history must not open a transaction")
		}
	}
	if err = db.Callback().Update().Before("gorm:update").Register("test:no_update_tx", check); err != nil {
		t.Fatal(err)
	}
	if err = db.Callback().Create().Before("gorm:create").Register("test:no_create_tx", check); err != nil {
		t.Fatal(err)
	}
	if err = db.Callback().Delete().Before("gorm:delete").Register("test:no_delete_tx", check); err != nil {
		t.Fatal(err)
	}
	for v := uint64(0); v < 15; v++ {
		post.Content = fmt.Sprintf("body %d", v)
		post.ModerationVersion = v + 1
		if err = SaveReviewedWithRevisionWithDB(db, post, v); err != nil {
			t.Fatal(err)
		}
	}
	post.Content = "stale"
	if err = SaveReviewedWithRevisionWithDB(db, post, 0); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("conflict: %v", err)
	}
	page, err := postRevisions.ListPageWithDB(db, postRevisions.PageQuery{PostId: post.Id})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 10 || page.Items[0].RevisionNo != 15 || page.Items[9].RevisionNo != 6 {
		t.Fatalf("history: %+v", page)
	}
	// An unchanged body must not add history, even when the moderation version changes.
	post.Content = "body 14"
	post.ModerationVersion = 16
	if err = SaveReviewedWithRevisionWithDB(db, post, 15); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err = db.Model(&postRevisions.Entity{}).Where("post_id = ?", post.Id).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 10 {
		t.Fatalf("unchanged body history count: %d", count)
	}
}

func TestHistoryVersionGapsAndDelayedWrites(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&postRevisions.Entity{}); err != nil {
		t.Fatal(err)
	}
	// Existing history used consecutive numbers, independent of moderation.
	if err = db.Create(&postRevisions.Entity{PostId: 7, RevisionNo: 1, Content: "legacy"}).Error; err != nil {
		t.Fatal(err)
	}
	current := Entity{Id: 7, Content: "current"}
	// Complete newer snapshots first. Old completions must not displace them.
	for version := uint64(30); version >= 10; version-- {
		RecordRevisionBestEffort(db, Entity{Id: 7, Content: fmt.Sprintf("old %d", version), ModerationVersion: version}, current)
	}
	// A duplicate callback is best-effort and must not replace immutable content.
	RecordRevisionBestEffort(db, Entity{Id: 7, Content: "duplicate", ModerationVersion: 30}, current)
	page, err := postRevisions.ListPageWithDB(db, postRevisions.PageQuery{PostId: 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 10 || page.Items[0].RevisionNo != 31 || page.Items[9].RevisionNo != 22 || page.Items[0].Content != "old 30" {
		t.Fatalf("history: %+v", page)
	}
}
