package postRevisions_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"gorm.io/gorm"
)

func TestNextRevisionNoWithDBSerializesAgainstPostRow(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&posts.Entity{}, &postRevisions.Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	post := &posts.Entity{TopicId: 1, PostNo: 1, UserId: 2, Content: "current"}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	if err := db.Create(&postRevisions.Entity{PostId: post.Id, RevisionNo: 1, Content: "old"}).Error; err != nil {
		t.Fatalf("create revision: %v", err)
	}
	var got uint64
	if err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		got, err = postRevisions.NextRevisionNoWithDB(tx, post.Id)
		return err
	}); err != nil {
		t.Fatalf("next revision: %v", err)
	}
	if got != 2 {
		t.Fatalf("next revision = %d, want 2", got)
	}
}

func TestNextRevisionNoWithDBRejectsMissingPost(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&posts.Entity{}, &postRevisions.Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := postRevisions.NextRevisionNoWithDB(db, 99); err == nil {
		t.Fatal("missing post should return an error")
	}
}
