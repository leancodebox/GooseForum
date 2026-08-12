package accessadminservice

import (
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
)

func TestRestrictingCategoryDoesNotTouchTopics(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(
		&accessGroups.Entity{}, &accessGroupMembers.Entity{}, &categoryGroupPermissions.Entity{},
		&category.Entity{}, &topics.Entity{}, &topicCategoryIndex.Entity{},
	); err != nil {
		t.Fatalf("migrate access admin tables: %v", err)
	}
	const restrictedCategoryID uint64 = 960001
	const otherCategoryID uint64 = 960002
	const topicID uint64 = 960101
	if err := conn.Create(&[]category.Entity{{Id: restrictedCategoryID, Name: "Private"}, {Id: otherCategoryID, Name: "Other"}}).Error; err != nil {
		t.Fatalf("create categories: %v", err)
	}
	result := datamigration.BackfillAccessControlDefaultsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("backfill defaults: %s", result.LastFailed)
	}
	groups, err := accessGroups.GetBySystemKeys([]string{accessGroups.SystemKeyEveryone, accessGroups.SystemKeyRegistered})
	if err != nil {
		t.Fatalf("load system groups: %v", err)
	}
	groupID := func(key string) uint64 {
		for _, group := range groups {
			if group.SystemKey != nil && *group.SystemKey == key {
				return group.Id
			}
		}
		return 0
	}
	everyoneID := groupID(accessGroups.SystemKeyEveryone)
	registeredID := groupID(accessGroups.SystemKeyRegistered)
	topic := topics.Entity{Id: topicID, Title: "multi category", CategoryIds: []uint64{restrictedCategoryID, otherCategoryID}, MainCategoryId: restrictedCategoryID, UserId: 7, Status: 1}
	if err := conn.Create(&topic).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if err := conn.Create(&[]topicCategoryIndex.Entity{
		{TopicId: topicID, CategoryId: restrictedCategoryID, Effective: 1},
		{TopicId: topicID, CategoryId: otherCategoryID, Effective: 1},
	}).Error; err != nil {
		t.Fatalf("create category indexes: %v", err)
	}
	// Making a public category private used to require rewriting every
	// multi-category topic in it. Visibility now comes from main_category_id,
	// which no permission change can invalidate, so there is nothing to convert
	// and no strategy to ask the admin for.
	grants := []GrantInput{{AccessGroupID: everyoneID, Level: 0}, {AccessGroupID: registeredID, Level: categoryGroupPermissions.PermissionCreate}}
	if err := ReplaceCategoryGrants(restrictedCategoryID, grants); err != nil {
		t.Fatalf("replace grants: %v", err)
	}
	unchanged := topics.GetSimple(topicID)
	if !reflect.DeepEqual(unchanged.CategoryIds, []uint64{restrictedCategoryID, otherCategoryID}) {
		t.Fatalf("topic categories mutated: %v", unchanged.CategoryIds)
	}
	if unchanged.MainCategoryId != restrictedCategoryID {
		t.Fatalf("main category mutated: %d", unchanged.MainCategoryId)
	}
	indexes := topicCategoryIndex.GetByTopicId(topicID)
	active := make(map[uint64]int)
	for _, index := range indexes {
		active[index.CategoryId] = index.Effective
	}
	if active[restrictedCategoryID] != 1 || active[otherCategoryID] != 1 {
		t.Fatalf("category indexes mutated: %v", active)
	}
	accesscontrol.InvalidateSystemGroups()
	for _, group := range groups {
		accesscontrol.InvalidateGroup(group.Id)
	}
	guest, err := accesscontrol.Resolve(0)
	if err != nil {
		t.Fatalf("resolve guest: %v", err)
	}
	if guest.CanReadCategory(restrictedCategoryID) {
		t.Fatal("restricted category remained guest-readable")
	}
}

func TestApplicationOnlyActivatesMembershipAfterApproval(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&accessGroups.Entity{}, &accessGroupMembers.Entity{}); err != nil {
		t.Fatalf("migrate application tables: %v", err)
	}
	group := accessGroups.Entity{Id: 961001, Name: "Applicants", JoinMode: accessGroups.JoinModeApplication, Status: accessGroups.StatusEnabled}
	if err := conn.Create(&group).Error; err != nil {
		t.Fatalf("create application group: %v", err)
	}
	const userID uint64 = 961010
	if err := ApplyToGroup(group.Id, userID); err != nil {
		t.Fatalf("apply to group: %v", err)
	}
	member, err := accessGroupMembers.GetByGroupUser(group.Id, userID)
	if err != nil || member.Status != accessGroupMembers.StatusPending {
		t.Fatalf("pending member = %+v, %v", member, err)
	}
	active, err := accessGroupMembers.ActiveGroupIDsByUser(userID)
	if err != nil || len(active) != 0 {
		t.Fatalf("active groups before review = %v, %v", active, err)
	}
	if err := ReviewApplication(group.Id, member.Id, true); err != nil {
		t.Fatalf("approve application: %v", err)
	}
	active, err = accessGroupMembers.ActiveGroupIDsByUser(userID)
	if err != nil || !reflect.DeepEqual(active, []uint64{group.Id}) {
		t.Fatalf("active groups after review = %v, %v", active, err)
	}
}

func TestCategoryLifecycleKeepsDefaultGrantsTransactional(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&accessGroups.Entity{}, &categoryGroupPermissions.Entity{}, &category.Entity{}); err != nil {
		t.Fatalf("migrate category access tables: %v", err)
	}
	result := datamigration.BackfillAccessControlDefaultsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("ensure system groups: %s", result.LastFailed)
	}
	entity := category.Entity{Id: 962001, Name: "New category"}
	if err := SaveCategoryWithDefaults(&entity, true); err != nil {
		t.Fatalf("save category defaults: %v", err)
	}
	grants, err := categoryGroupPermissions.ByCategoryIDs([]uint64{entity.Id})
	if err != nil || len(grants) != 2 {
		t.Fatalf("new category grants = %+v, %v", grants, err)
	}
	if err := DeleteCategory(&entity); err != nil {
		t.Fatalf("delete category: %v", err)
	}
	grants, err = categoryGroupPermissions.ByCategoryIDs([]uint64{entity.Id})
	if err != nil || len(grants) != 0 {
		t.Fatalf("deleted category grants = %+v, %v", grants, err)
	}
}

func TestDisablingGroupImmediatelyRevokesCachedMembership(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(
		&accessGroups.Entity{}, &accessGroupMembers.Entity{}, &categoryGroupPermissions.Entity{}, &category.Entity{},
	); err != nil {
		t.Fatalf("migrate group revocation tables: %v", err)
	}
	result := datamigration.BackfillAccessControlDefaultsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("ensure system groups: %s", result.LastFailed)
	}
	group := accessGroups.Entity{Id: 963001, Name: "Temporary", JoinMode: accessGroups.JoinModeInviteOnly, Status: accessGroups.StatusEnabled}
	if err := conn.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	const userID uint64 = 963010
	member := accessGroupMembers.Entity{AccessGroupId: group.Id, UserId: userID, MemberRole: accessGroupMembers.MemberRoleMember, Status: accessGroupMembers.StatusEnabled}
	if err := conn.Create(&member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	grant := categoryGroupPermissions.Entity{CategoryId: 963020, AccessGroupId: group.Id, PermissionLevel: categoryGroupPermissions.PermissionRead, Status: categoryGroupPermissions.StatusEnabled}
	if err := conn.Create(&grant).Error; err != nil {
		t.Fatalf("create grant: %v", err)
	}
	accesscontrol.InvalidateSystemGroups()
	accesscontrol.InvalidateUser(userID)
	accesscontrol.InvalidateGroup(group.Id)
	before, err := accesscontrol.Resolve(userID)
	if err != nil || !before.CanReadCategory(grant.CategoryId) {
		t.Fatalf("resolve before disable: readable=%v err=%v", before.CanReadCategory(grant.CategoryId), err)
	}

	if _, err := SaveGroup(GroupInput{ID: group.Id, Name: group.Name, JoinMode: group.JoinMode, Status: accessGroups.StatusDisabled}); err != nil {
		t.Fatalf("disable group: %v", err)
	}
	after, err := accesscontrol.Resolve(userID)
	if err != nil {
		t.Fatalf("resolve after disable: %v", err)
	}
	if after.CanReadCategory(grant.CategoryId) {
		t.Fatal("disabled group grant remained readable through cached membership")
	}
}
