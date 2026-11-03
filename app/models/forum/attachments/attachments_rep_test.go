package attachments

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestBatchBindingQueryBudgetAndPermanentClaims(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	files := make([]Info, 20)
	for i := range files {
		row := Entity{UserID: 7, ObjectName: fmt.Sprintf("batch-%d", i), Filename: "a.txt"}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
		files[i] = Info{ID: row.ID}
	}
	queries, updates := 0, 0
	if err := db.Callback().Query().Before("gorm:query").Register("binding-query-count", func(tx *gorm.DB) {
		if tx.Statement.Table == "attachments" {
			queries++
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Update().Before("gorm:update").Register("binding-update-count", func(tx *gorm.DB) {
		if tx.Statement.Table == "attachments" {
			updates++
		}
	}); err != nil {
		t.Fatal(err)
	}
	bind := func(userID, postID uint64, list []Info) error {
		return db.Table("posts").Transaction(func(tx *gorm.DB) error { return BindWithDB(tx, userID, postID, list) })
	}
	if err := bind(7, 31, files); err != nil {
		t.Fatal(err)
	}
	if queries != 1 || updates != 1 {
		t.Fatalf("queries=%d updates=%d, want 1 each for 20 attachments", queries, updates)
	}
	queries, updates = 0, 0
	if err := bind(7, 31, files); err != nil {
		t.Fatal(err)
	}
	if queries != 1 || updates != 0 {
		t.Fatalf("existing claims: queries=%d updates=%d", queries, updates)
	}
	for _, tc := range []struct {
		user, post uint64
		files      []Info
	}{
		{8, 31, files}, {7, 32, files},
		{7, 31, []Info{files[0], files[0]}},
		{7, 31, []Info{files[0], {ID: 999999}}},
	} {
		if err := bind(tc.user, tc.post, tc.files); err == nil {
			t.Fatal("invalid claim accepted")
		}
	}
}

func TestBatchBindingRollsBackWhenClaimCountChanges(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	rows := []Entity{{UserID: 7, ObjectName: "claim-first"}, {UserID: 7, ObjectName: "claim-second"}}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	// Simulate a claim disappearing between the locked read and conditional update.
	if err := db.Callback().Update().Before("gorm:update").Register("interrupt-batch-claim", func(tx *gorm.DB) {
		if tx.Statement.Table == "attachments" {
			tx.AddError(tx.Exec("DELETE FROM attachments WHERE id = ?", rows[1].ID).Error)
		}
	}); err != nil {
		t.Fatal(err)
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		return BindWithDB(tx, 7, 31, []Info{{ID: rows[0].ID}, {ID: rows[1].ID}})
	})
	if err != ErrInvalid {
		t.Fatalf("claim count mismatch: %v", err)
	}
	var after []Entity
	if err := db.Order("id").Find(&after).Error; err != nil {
		t.Fatal(err)
	}
	if len(after) != 2 || after[0].PostID != 0 || after[1].PostID != 0 {
		t.Fatalf("partial binding escaped rollback: %+v", after)
	}
}
