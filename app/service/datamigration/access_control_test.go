package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"gorm.io/gorm"
)

func TestBackfillAccessControlDefaultsIsCompatibleAndIdempotent(t *testing.T) {
	conn := openAccessControlMigrationDB(t)
	if err := conn.Create(&[]category.Entity{
		{Id: 880001, Name: "Public A"},
		{Id: 880002, Name: "Public B"},
	}).Error; err != nil {
		t.Fatalf("create categories: %v", err)
	}

	first := BackfillAccessControlDefaultsWithDB(conn)
	if first.Failed != 0 || first.Groups != 2 || first.Categories != 2 || first.Grants != 4 {
		t.Fatalf("first migration = %+v", first)
	}

	groups := loadSystemGroupsForTest(t, conn)
	assertGrantLevel(t, conn, 880001, groups[accessGroups.SystemKeyEveryone], categoryGroupPermissions.PermissionRead)
	assertGrantLevel(t, conn, 880001, groups[accessGroups.SystemKeyRegistered], categoryGroupPermissions.PermissionCreate)

	if err := conn.Model(&categoryGroupPermissions.Entity{}).
		Where("category_id = ? AND access_group_id = ?", 880001, groups[accessGroups.SystemKeyRegistered]).
		Update("permission_level", categoryGroupPermissions.PermissionReply).Error; err != nil {
		t.Fatalf("customize compatibility grant: %v", err)
	}

	second := BackfillAccessControlDefaultsWithDB(conn)
	if second.Failed != 0 {
		t.Fatalf("second migration = %+v", second)
	}
	assertGrantLevel(t, conn, 880001, groups[accessGroups.SystemKeyRegistered], categoryGroupPermissions.PermissionReply)

	var groupCount int64
	if err := conn.Model(&accessGroups.Entity{}).Count(&groupCount).Error; err != nil {
		t.Fatalf("count groups: %v", err)
	}
	if groupCount != 2 {
		t.Fatalf("group count after rerun = %d, want 2", groupCount)
	}
	var grantCount int64
	if err := conn.Model(&categoryGroupPermissions.Entity{}).Count(&grantCount).Error; err != nil {
		t.Fatalf("count grants: %v", err)
	}
	if grantCount != 4 {
		t.Fatalf("grant count after rerun = %d, want 4", grantCount)
	}
}

func openAccessControlMigrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(
		&accessGroups.Entity{},
		&accessGroupMembers.Entity{},
		&categoryGroupPermissions.Entity{},
		&category.Entity{},
	); err != nil {
		t.Fatalf("migrate access control schema: %v", err)
	}
	return conn
}

func loadSystemGroupsForTest(t *testing.T, conn *gorm.DB) map[string]uint64 {
	t.Helper()
	var groups []accessGroups.Entity
	if err := conn.Order("id ASC").Find(&groups).Error; err != nil {
		t.Fatalf("load groups: %v", err)
	}
	result := make(map[string]uint64, len(groups))
	for _, group := range groups {
		if group.SystemKey != nil {
			result[*group.SystemKey] = group.Id
		}
	}
	return result
}

func assertGrantLevel(t *testing.T, conn *gorm.DB, categoryID uint64, groupID uint64, want int8) {
	t.Helper()
	var grant categoryGroupPermissions.Entity
	if err := conn.Where("category_id = ? AND access_group_id = ?", categoryID, groupID).First(&grant).Error; err != nil {
		t.Fatalf("load grant category=%d group=%d: %v", categoryID, groupID, err)
	}
	if grant.PermissionLevel != want {
		t.Fatalf("grant category=%d group=%d level=%d, want %d", categoryID, groupID, grant.PermissionLevel, want)
	}
}
