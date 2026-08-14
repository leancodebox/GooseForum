package forum

import (
	"net/http"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/moderationLog"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
)

func TestAccessGroupManageGrantCanModerateCategoryContent(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(
		&accessGroups.Entity{}, &accessGroupMembers.Entity{}, &categoryGroupPermissions.Entity{},
		&category.Entity{}, &topics.Entity{}, &moderationLog.Entity{},
	); err != nil {
		t.Fatalf("migrate moderation access tables: %v", err)
	}
	const categoryID uint64 = 975001
	const groupID uint64 = 975002
	const userID uint64 = 975003
	const topicID uint64 = 975004
	const otherCategoryID uint64 = 975006
	const otherTopicID uint64 = 975007
	conn.Unscoped().Delete(&topics.Entity{}, []uint64{topicID, otherTopicID})
	conn.Unscoped().Delete(&category.Entity{}, []uint64{categoryID, otherCategoryID})
	conn.Unscoped().Delete(&accessGroups.Entity{}, groupID)
	conn.Where("access_group_id = ? OR user_id = ?", groupID, userID).Delete(&accessGroupMembers.Entity{})
	conn.Where("category_id = ? OR access_group_id = ?", categoryID, groupID).Delete(&categoryGroupPermissions.Entity{})

	if err := conn.Create(&[]category.Entity{{Id: categoryID, Name: "Managed"}, {Id: otherCategoryID, Name: "Outside scope"}}).Error; err != nil {
		t.Fatalf("create categories: %v", err)
	}
	result := datamigration.BackfillAccessControlDefaultsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("backfill access defaults: %s", result.LastFailed)
	}
	group := accessGroups.Entity{Id: groupID, Name: "Category managers", JoinMode: accessGroups.JoinModeInviteOnly, Status: accessGroups.StatusEnabled}
	if err := conn.Create(&group).Error; err != nil {
		t.Fatalf("create access group: %v", err)
	}
	if err := conn.Create(&accessGroupMembers.Entity{AccessGroupId: groupID, UserId: userID, MemberRole: accessGroupMembers.MemberRoleMember, Status: accessGroupMembers.StatusEnabled}).Error; err != nil {
		t.Fatalf("create group member: %v", err)
	}
	if err := conn.Create(&categoryGroupPermissions.Entity{CategoryId: categoryID, AccessGroupId: groupID, PermissionLevel: categoryGroupPermissions.PermissionManage, Status: categoryGroupPermissions.StatusEnabled}).Error; err != nil {
		t.Fatalf("create manage grant: %v", err)
	}
	topic := topics.Entity{Id: topicID, Title: "Managed topic", UserId: 975005, CategoryIds: []uint64{categoryID}, MainCategoryId: categoryID, Status: 1}
	if err := conn.Create(&topic).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}
	otherTopic := topics.Entity{Id: otherTopicID, Title: "Outside topic", UserId: 975005, CategoryIds: []uint64{otherCategoryID}, MainCategoryId: otherCategoryID, Status: 1}
	if err := conn.Create(&otherTopic).Error; err != nil {
		t.Fatalf("create outside topic: %v", err)
	}
	accesscontrol.InvalidateSystemGroups()
	accesscontrol.InvalidateGroup(groupID)
	accesscontrol.InvalidateUser(userID)

	global, categoryIDs, ok := moderationAccessScope(userID)
	if !ok || global || len(categoryIDs) != 1 || categoryIDs[0] != categoryID {
		t.Fatalf("moderation scope = global:%v categories:%v ok:%v", global, categoryIDs, ok)
	}
	response := UpdateModerationTopicStatus(component.BetterRequest[ModerationTopicStatusReq]{
		UserId: userID,
		Params: ModerationTopicStatusReq{TopicId: topicID, Action: "ban"},
	})
	if response.Code != http.StatusOK || response.Data.Code != component.SUCCESS {
		t.Fatalf("moderation response = %#v", response)
	}
	if got := topics.GetSimple(topicID); got.ProcessStatus != 1 {
		t.Fatalf("topic process status = %d, want 1", got.ProcessStatus)
	}
	logRecord := moderationLog.Entity{
		ActorUserId: userID,
		Action:      moderationLog.ActionTopicBlocked,
		SubjectType: moderationLog.SubjectTopic,
		SubjectId:   topicID,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.topic.blocked",
			Params:      map[string]any{"topicId": topicID, "title": topic.Title},
		},
	}
	if err := moderationLog.Create(&logRecord); err != nil {
		t.Fatalf("create scoped moderation log: %v", err)
	}
	outsideLog := moderationLog.Entity{
		ActorUserId: userID,
		Action:      moderationLog.ActionTopicBlocked,
		SubjectType: moderationLog.SubjectTopic,
		SubjectId:   otherTopicID,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.topic.blocked",
			Params:      map[string]any{"topicId": otherTopicID, "title": otherTopic.Title},
		},
	}
	if err := moderationLog.Create(&outsideLog); err != nil {
		t.Fatalf("create outside moderation log: %v", err)
	}
	logResponse := ModerationLogList(component.BetterRequest[ModerationLogListReq]{UserId: userID, Params: ModerationLogListReq{PageSize: 10}})
	logPage, ok := logResponse.Data.Result.(ModerationLogListResponse)
	if logResponse.Code != http.StatusOK || !ok || len(logPage.Items) == 0 {
		t.Fatalf("moderation log response = %#v", logResponse)
	}
	found := false
	for _, item := range logPage.Items {
		if item.Subject.ID == otherTopicID {
			t.Fatalf("outside category log leaked into managed scope: %#v", item)
		}
		if item.Subject.ID == topicID && item.Action == moderationLog.ActionTopicBlocked {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("managed category log missing from %#v", logPage.Items)
	}
}
