package fileaccessservice

import (
	"fmt"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/moderators"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
)

func TestResolveProtectsRestrictedTopicFilesAndPendingUploads(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(
		&accessGroups.Entity{}, &accessGroupMembers.Entity{}, &categoryGroupPermissions.Entity{},
		&category.Entity{}, &topics.Entity{}, &posts.Entity{}, &fileUsage.Entity{},
		&users.EntityComplete{}, &moderators.Entity{},
	); err != nil {
		t.Fatalf("migrate file access tables: %v", err)
	}
	publicCategory := category.Entity{Id: 970001, Name: "Public"}
	restrictedCategory := category.Entity{Id: 970002, Name: "Restricted"}
	if err := conn.Create(&[]category.Entity{publicCategory, restrictedCategory}).Error; err != nil {
		t.Fatalf("create categories: %v", err)
	}
	result := datamigration.BackfillAccessControlDefaultsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("backfill access defaults: %s", result.LastFailed)
	}
	system, err := accessGroups.GetBySystemKeys([]string{accessGroups.SystemKeyEveryone, accessGroups.SystemKeyRegistered})
	if err != nil {
		t.Fatalf("load system groups: %v", err)
	}
	for _, group := range system {
		if err := conn.Model(&categoryGroupPermissions.Entity{}).
			Where("category_id = ? AND access_group_id = ?", restrictedCategory.Id, group.Id).
			Update("status", categoryGroupPermissions.StatusDisabled).Error; err != nil {
			t.Fatalf("disable restricted system grant: %v", err)
		}
	}
	custom := accessGroups.Entity{Name: "Private members", JoinMode: accessGroups.JoinModeInviteOnly, Status: accessGroups.StatusEnabled}
	if err := conn.Create(&custom).Error; err != nil {
		t.Fatalf("create custom group: %v", err)
	}
	if err := conn.Create(&categoryGroupPermissions.Entity{CategoryId: restrictedCategory.Id, AccessGroupId: custom.Id, PermissionLevel: categoryGroupPermissions.PermissionRead, Status: categoryGroupPermissions.StatusEnabled}).Error; err != nil {
		t.Fatalf("create restricted grant: %v", err)
	}
	if err := conn.Create(&accessGroupMembers.Entity{AccessGroupId: custom.Id, UserId: 970010, MemberRole: accessGroupMembers.MemberRoleMember, Status: accessGroupMembers.StatusEnabled}).Error; err != nil {
		t.Fatalf("create private member: %v", err)
	}
	for _, userID := range []uint64{970010, 970011, 970012} {
		if err := conn.Create(&users.EntityComplete{Id: userID, Username: fmt.Sprintf("file-user-%d", userID)}).Error; err != nil {
			t.Fatalf("create user %d: %v", userID, err)
		}
	}
	publicTopic := topics.Entity{Id: 970101, Title: "public", CategoryIds: []uint64{publicCategory.Id}, MainCategoryId: publicCategory.Id, UserId: 970012, Status: 1}
	restrictedTopic := topics.Entity{Id: 970102, Title: "private", CategoryIds: []uint64{restrictedCategory.Id}, MainCategoryId: restrictedCategory.Id, UserId: 970012, Status: 1}
	if err := conn.Create(&[]topics.Entity{publicTopic, restrictedTopic}).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}
	if err := conn.Create(&[]fileUsage.Entity{
		{FileName: "public.webp", TargetType: fileUsage.TargetTopic, TargetId: publicTopic.Id, UsageType: fileUsage.UsageInlineImage, UserId: publicTopic.UserId},
		{FileName: "private.webp", TargetType: fileUsage.TargetTopic, TargetId: restrictedTopic.Id, UsageType: fileUsage.UsageInlineImage, UserId: restrictedTopic.UserId},
		{FileName: "pending.webp", TargetType: fileUsage.TargetPendingUpload, TargetId: 970012, UsageType: fileUsage.UsagePendingUpload, UserId: 970012},
	}).Error; err != nil {
		t.Fatalf("create file usages: %v", err)
	}
	accesscontrol.InvalidateSystemGroups()
	for _, group := range append(system, custom) {
		accesscontrol.InvalidateGroup(group.Id)
	}
	accesscontrol.InvalidateUser(970010)
	accesscontrol.InvalidateUser(970011)

	assertDecision(t, 0, "public.webp", 970012, Decision{Allowed: true, Public: true})
	assertDecision(t, 0, "private.webp", 970012, Decision{})
	assertDecision(t, 970010, "private.webp", 970012, Decision{Allowed: true})
	assertDecision(t, 970011, "private.webp", 970012, Decision{})
	assertDecision(t, 970012, "pending.webp", 970012, Decision{Allowed: true})
	assertDecision(t, 970011, "pending.webp", 970012, Decision{})
}

func assertDecision(t *testing.T, userID uint64, fileName string, uploaderID uint64, want Decision) {
	t.Helper()
	got, err := Resolve(userID, fileName, uploaderID)
	if err != nil {
		t.Fatalf("Resolve(%d, %q): %v", userID, fileName, err)
	}
	if got != want {
		t.Fatalf("Resolve(%d, %q) = %+v, want %+v", userID, fileName, got, want)
	}
}
