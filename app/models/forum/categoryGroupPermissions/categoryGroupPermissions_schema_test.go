package categoryGroupPermissions

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"gorm.io/gorm"
)

func TestAccessControlSchemaMigratesWithRequiredIndexes(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(
		&accessGroups.Entity{},
		&accessGroupMembers.Entity{},
		&Entity{},
	); err != nil {
		t.Fatalf("migrate access control schema: %v", err)
	}

	for _, table := range []string{"access_groups", "access_group_members", "category_group_permissions"} {
		if !conn.Migrator().HasTable(table) {
			t.Fatalf("%s table was not created", table)
		}
	}
	checks := []struct {
		model any
		index string
	}{
		{&accessGroups.Entity{}, "uniq_access_group_system_key"},
		{&accessGroupMembers.Entity{}, "uniq_access_group_member"},
		{&accessGroupMembers.Entity{}, "idx_access_group_members_user_status_group"},
		{&Entity{}, "uniq_category_group_permission"},
		{&Entity{}, "idx_category_group_permissions_group_status_category"},
	}
	for _, check := range checks {
		if !conn.Migrator().HasIndex(check.model, check.index) {
			t.Fatalf("index %s was not created", check.index)
		}
	}

	customA := accessGroups.Entity{Name: "A"}
	customB := accessGroups.Entity{Name: "B"}
	if err := conn.Create(&customA).Error; err != nil {
		t.Fatalf("create first custom group with null system key: %v", err)
	}
	if err := conn.Create(&customB).Error; err != nil {
		t.Fatalf("create second custom group with null system key: %v", err)
	}
}
