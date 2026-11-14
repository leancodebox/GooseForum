package topics

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"gorm.io/gorm"
)

func TestPublicMaxHonorsMainCategoryAndPublication(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&Entity{}, &accessGroups.Entity{}, &category.Entity{}, &categoryGroupPermissions.Entity{}); err != nil {
		t.Fatal(err)
	}
	everyone := accessGroups.SystemKeyEveryone
	for _, row := range []any{
		&category.Entity{Id: 1}, &category.Entity{Id: 2},
		&accessGroups.Entity{Id: 1, SystemKey: &everyone, Status: 1},
		&categoryGroupPermissions.Entity{CategoryId: 1, AccessGroupId: 1, PermissionLevel: 1, Status: 1},
		&Entity{Id: 10, MainCategoryId: 1, Status: 1},
		&Entity{Id: 20, MainCategoryId: 2, CategoryIds: []uint64{2, 1}, Status: 1},
		&Entity{Id: 30, MainCategoryId: 1, Status: 0},
		&Entity{Id: 40, MainCategoryId: 1, Status: 1, ProcessStatus: 1},
		&Entity{Id: 50, MainCategoryId: 1, Status: 1},
	} {
		if err = db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err = db.Delete(&Entity{}, 50).Error; err != nil {
		t.Fatal(err)
	}
	assertMax := func(want uint64) {
		t.Helper()
		id, err := PublicMaxIDWithDB(db)
		if err != nil || id != want {
			t.Fatalf("max=%d err=%v want=%d", id, err, want)
		}
	}
	assertMax(10)
	if err = db.Model(&categoryGroupPermissions.Entity{}).Where("category_id = ?", 1).Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	assertMax(0)
	if err = db.Model(&categoryGroupPermissions.Entity{}).Where("category_id = ?", 1).Update("status", 1).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Model(&Entity{}).Where("id = ?", 30).Update("status", 1).Error; err != nil {
		t.Fatal(err)
	}
	assertMax(30)
}
