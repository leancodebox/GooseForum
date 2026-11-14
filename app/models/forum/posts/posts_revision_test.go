package posts

import (
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

func TestSaveReviewedWithRevisionRollsBackWhenHistoryInsertCannotRun(t *testing.T) {
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
	if err := SaveReviewedWithRevisionWithDB(db, post, 0); err == nil {
		t.Fatal("missing history table should fail")
	}
	var stored Entity
	if err := db.First(&stored, post.Id).Error; err != nil {
		t.Fatalf("load post after rollback: %v", err)
	}
	if stored.Content != "before" || stored.ModerationVersion != 0 {
		t.Fatalf("post after rollback = %+v, want original content and version", stored)
	}
}
