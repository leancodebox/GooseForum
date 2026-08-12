package accessGroupMembers

import (
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
)

func TestActiveUserIDsWithCategoryCapabilityFiltersWholeBatch(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&accessGroups.Entity{}, &Entity{}, &categoryGroupPermissions.Entity{}); err != nil {
		t.Fatalf("migrate access control tables: %v", err)
	}
	const categoryID = uint64(974001)
	groupIDs := []uint64{974011, 974012, 974013}
	userIDs := []uint64{974021, 974022, 974023, 974024}
	conn.Unscoped().Where("access_group_id IN ?", groupIDs).Delete(&categoryGroupPermissions.Entity{})
	conn.Unscoped().Where("access_group_id IN ?", groupIDs).Delete(&Entity{})
	conn.Unscoped().Where("id IN ?", groupIDs).Delete(&accessGroups.Entity{})
	t.Cleanup(func() {
		conn.Unscoped().Where("access_group_id IN ?", groupIDs).Delete(&categoryGroupPermissions.Entity{})
		conn.Unscoped().Where("access_group_id IN ?", groupIDs).Delete(&Entity{})
		conn.Unscoped().Where("id IN ?", groupIDs).Delete(&accessGroups.Entity{})
	})

	groups := []accessGroups.Entity{
		{Id: groupIDs[0], Name: "enabled", Status: accessGroups.StatusEnabled},
		{Id: groupIDs[1], Name: "disabled", Status: accessGroups.StatusDisabled},
		{Id: groupIDs[2], Name: "disabled grant", Status: accessGroups.StatusEnabled},
	}
	if err := conn.Create(&groups).Error; err != nil {
		t.Fatalf("create groups: %v", err)
	}
	if err := conn.Model(&accessGroups.Entity{}).Where("id = ?", groupIDs[1]).Update("status", accessGroups.StatusDisabled).Error; err != nil {
		t.Fatalf("disable group: %v", err)
	}
	members := []Entity{
		{AccessGroupId: groupIDs[0], UserId: userIDs[0], Status: StatusEnabled},
		{AccessGroupId: groupIDs[0], UserId: userIDs[1], Status: StatusPending},
		{AccessGroupId: groupIDs[1], UserId: userIDs[2], Status: StatusEnabled},
		{AccessGroupId: groupIDs[2], UserId: userIDs[3], Status: StatusEnabled},
	}
	if err := conn.Create(&members).Error; err != nil {
		t.Fatalf("create members: %v", err)
	}
	grants := []categoryGroupPermissions.Entity{
		{CategoryId: categoryID, AccessGroupId: groupIDs[0], PermissionLevel: categoryGroupPermissions.PermissionRead, Status: categoryGroupPermissions.StatusEnabled},
		{CategoryId: categoryID, AccessGroupId: groupIDs[1], PermissionLevel: categoryGroupPermissions.PermissionRead, Status: categoryGroupPermissions.StatusEnabled},
		{CategoryId: categoryID, AccessGroupId: groupIDs[2], PermissionLevel: categoryGroupPermissions.PermissionRead, Status: categoryGroupPermissions.StatusDisabled},
	}
	if err := conn.Create(&grants).Error; err != nil {
		t.Fatalf("create grants: %v", err)
	}
	if err := conn.Model(&categoryGroupPermissions.Entity{}).Where("access_group_id = ?", groupIDs[2]).Update("status", categoryGroupPermissions.StatusDisabled).Error; err != nil {
		t.Fatalf("disable grant: %v", err)
	}

	got, err := ActiveUserIDsWithCategoryCapability(userIDs, categoryID, categoryGroupPermissions.PermissionRead)
	if err != nil {
		t.Fatalf("filter user ids: %v", err)
	}
	if want := []uint64{userIDs[0]}; !reflect.DeepEqual(got, want) {
		t.Fatalf("readable user ids = %v, want %v", got, want)
	}
}
