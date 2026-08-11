package api

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accessadminservice"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/samber/lo"
)

type AccessGroupMemberItem struct {
	ID         uint64 `json:"id"`
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatarUrl"`
	MemberRole string `json:"memberRole"`
	Status     int8   `json:"status"`
}

type AccessGroupGrantItem struct {
	CategoryID uint64 `json:"categoryId"`
	Level      int8   `json:"level"`
}

type AccessGroupItem struct {
	ID        uint64                  `json:"id"`
	Name      string                  `json:"name"`
	SystemKey string                  `json:"systemKey,omitempty"`
	JoinMode  string                  `json:"joinMode"`
	Status    int8                    `json:"status"`
	Members   []AccessGroupMemberItem `json:"members"`
	Grants    []AccessGroupGrantItem  `json:"grants"`
}

type AccessCategoryItem struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	IsRestricted bool   `json:"isRestricted"`
}

type AccessControlOverview struct {
	Groups     []AccessGroupItem    `json:"groups"`
	Categories []AccessCategoryItem `json:"categories"`
}

func GetAccessControlOverview(req component.BetterRequest[struct{}]) component.Response {
	groups, err := accessGroups.All()
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	members, err := accessGroupMembers.All()
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	members = lo.Filter(members, func(member accessGroupMembers.Entity, _ int) bool {
		return member.Status != accessGroupMembers.StatusDisabled
	})
	categories := category.All()
	categoryIDs := lo.Map(categories, func(item *category.Entity, _ int) uint64 { return item.Id })
	grants, err := categoryGroupPermissions.ByCategoryIDs(categoryIDs)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	userIDs := lo.Uniq(lo.Map(members, func(item accessGroupMembers.Entity, _ int) uint64 { return item.UserId }))
	userMap := users.GetMapByIds(userIDs)
	membersByGroup := lo.GroupBy(members, func(item accessGroupMembers.Entity) uint64 { return item.AccessGroupId })
	grantsByGroup := lo.GroupBy(grants, func(item categoryGroupPermissions.Entity) uint64 { return item.AccessGroupId })
	everyoneID := uint64(0)
	for _, group := range groups {
		if group.SystemKey != nil && *group.SystemKey == accessGroups.SystemKeyEveryone {
			everyoneID = group.Id
		}
	}
	items := make([]AccessGroupItem, 0, len(groups))
	for _, group := range groups {
		systemKey := ""
		if group.SystemKey != nil {
			systemKey = *group.SystemKey
		}
		memberItems := make([]AccessGroupMemberItem, 0, len(membersByGroup[group.Id]))
		for _, member := range membersByGroup[group.Id] {
			username, avatarURL := "", ""
			if user := userMap[member.UserId]; user != nil {
				username, avatarURL = user.Username, user.GetWebAvatarUrl()
			}
			memberItems = append(memberItems, AccessGroupMemberItem{ID: member.Id, UserID: member.UserId, Username: username, AvatarURL: avatarURL, MemberRole: member.MemberRole, Status: member.Status})
		}
		grantItems := make([]AccessGroupGrantItem, 0, len(grantsByGroup[group.Id]))
		for _, grant := range grantsByGroup[group.Id] {
			level := int8(0)
			if grant.Status == categoryGroupPermissions.StatusEnabled {
				level = grant.PermissionLevel
			}
			grantItems = append(grantItems, AccessGroupGrantItem{CategoryID: grant.CategoryId, Level: level})
		}
		items = append(items, AccessGroupItem{ID: group.Id, Name: group.Name, SystemKey: systemKey, JoinMode: group.JoinMode, Status: group.Status, Members: memberItems, Grants: grantItems})
	}
	publicCategories := make(map[uint64]bool)
	for _, grant := range grantsByGroup[everyoneID] {
		publicCategories[grant.CategoryId] = grant.Status == categoryGroupPermissions.StatusEnabled && grant.PermissionLevel >= categoryGroupPermissions.PermissionRead
	}
	categoryItems := lo.Map(categories, func(item *category.Entity, _ int) AccessCategoryItem {
		return AccessCategoryItem{ID: item.Id, Name: item.Name, Color: item.Color, IsRestricted: !publicCategories[item.Id]}
	})
	return component.SuccessResponse(AccessControlOverview{Groups: items, Categories: categoryItems})
}

type SaveAccessGroupReq struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name" validate:"required"`
	JoinMode string `json:"joinMode" validate:"oneof=invite_only application"`
	Status   int8   `json:"status" validate:"oneof=0 1"`
}

func SaveAccessGroup(req component.BetterRequest[SaveAccessGroupReq]) component.Response {
	group, err := accessadminservice.SaveGroup(accessadminservice.GroupInput{ID: req.Params.ID, Name: req.Params.Name, JoinMode: req.Params.JoinMode, Status: req.Params.Status, CreatedBy: req.UserId})
	if err != nil {
		return accessAdminFailure("save access group", err)
	}
	return component.SuccessResponse(group.Id)
}

func DeleteAccessGroup(req component.BetterRequest[struct {
	ID uint64 `json:"id" validate:"required"`
}]) component.Response {
	if err := accessadminservice.DeleteGroup(req.Params.ID); err != nil {
		return accessAdminFailure("delete access group", err)
	}
	return component.SuccessResponse(true)
}

type SaveAccessGroupMemberReq struct {
	GroupID    uint64 `json:"groupId" validate:"required"`
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	MemberRole string `json:"memberRole" validate:"oneof=member manager"`
}

func SaveAccessGroupMember(req component.BetterRequest[SaveAccessGroupMemberReq]) component.Response {
	user, ok := resolveModeratorUser(ModeratorUserReq{UserId: req.Params.UserID, Username: strings.TrimSpace(req.Params.Username)})
	if !ok {
		return component.FailResponseCode(component.MessageUserNotFound, nil)
	}
	member, err := accessadminservice.SaveMember(req.Params.GroupID, user.Id, req.Params.MemberRole, req.UserId)
	if err != nil {
		return accessAdminFailure("save access group member", err)
	}
	return component.SuccessResponse(member.Id)
}

func DeleteAccessGroupMember(req component.BetterRequest[struct {
	GroupID  uint64 `json:"groupId" validate:"required"`
	MemberID uint64 `json:"memberId" validate:"required"`
}]) component.Response {
	if err := accessadminservice.DeleteMember(req.Params.GroupID, req.Params.MemberID); err != nil {
		return accessAdminFailure("delete access group member", err)
	}
	return component.SuccessResponse(true)
}

func ReviewAccessGroupApplication(req component.BetterRequest[struct {
	GroupID  uint64 `json:"groupId" validate:"required"`
	MemberID uint64 `json:"memberId" validate:"required"`
	Approve  bool   `json:"approve"`
}]) component.Response {
	if err := accessadminservice.ReviewApplication(req.Params.GroupID, req.Params.MemberID, req.Params.Approve); err != nil {
		return accessAdminFailure("review access group application", err)
	}
	return component.SuccessResponse(true)
}

type JoinableAccessGroupItem struct {
	ID         uint64   `json:"id"`
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	Status     int8     `json:"status"`
}

func ListJoinableAccessGroups(req component.BetterRequest[struct{}]) component.Response {
	groups, err := accessGroups.All()
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	categories := category.All()
	categoryIDs := lo.Map(categories, func(item *category.Entity, _ int) uint64 { return item.Id })
	categoryByID := lo.SliceToMap(categories, func(item *category.Entity) (uint64, *category.Entity) { return item.Id, item })
	grants, err := categoryGroupPermissions.ByCategoryIDs(categoryIDs)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	grantsByGroup := lo.GroupBy(grants, func(item categoryGroupPermissions.Entity) uint64 { return item.AccessGroupId })
	result := make([]JoinableAccessGroupItem, 0)
	for _, group := range groups {
		if group.SystemKey != nil || group.Status != accessGroups.StatusEnabled || group.JoinMode != accessGroups.JoinModeApplication {
			continue
		}
		member, _ := accessGroupMembers.GetByGroupUser(group.Id, req.UserId)
		categoryNames := make([]string, 0)
		for _, grant := range grantsByGroup[group.Id] {
			if grant.Status == categoryGroupPermissions.StatusEnabled && grant.PermissionLevel >= categoryGroupPermissions.PermissionRead && categoryByID[grant.CategoryId] != nil {
				categoryNames = append(categoryNames, categoryByID[grant.CategoryId].Name)
			}
		}
		result = append(result, JoinableAccessGroupItem{ID: group.Id, Name: group.Name, Categories: categoryNames, Status: member.Status})
	}
	return component.SuccessResponse(result)
}

func ApplyToAccessGroup(req component.BetterRequest[struct {
	GroupID uint64 `json:"groupId" validate:"required"`
}]) component.Response {
	if err := accessadminservice.ApplyToGroup(req.Params.GroupID, req.UserId); err != nil {
		return accessAdminFailure("apply to access group", err)
	}
	return component.SuccessResponse(true)
}

type ManagedAccessGroupItem struct {
	ID           uint64                  `json:"id"`
	Name         string                  `json:"name"`
	Applications []AccessGroupMemberItem `json:"applications"`
}

func ListManagedAccessGroups(req component.BetterRequest[struct{}]) component.Response {
	groupIDs, err := accessGroupMembers.ManagedGroupIDsByUser(req.UserId)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	result := make([]ManagedAccessGroupItem, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group, err := accessGroups.Get(groupID)
		if err != nil || group.Id == 0 {
			continue
		}
		members, err := accessGroupMembers.ByGroupID(groupID)
		if err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		pending := lo.Filter(members, func(member accessGroupMembers.Entity, _ int) bool {
			return member.Status == accessGroupMembers.StatusPending
		})
		userIDs := lo.Map(pending, func(member accessGroupMembers.Entity, _ int) uint64 { return member.UserId })
		userMap := users.GetMapByIds(userIDs)
		items := make([]AccessGroupMemberItem, 0, len(pending))
		for _, member := range pending {
			username, avatarURL := "", ""
			if user := userMap[member.UserId]; user != nil {
				username, avatarURL = user.Username, user.GetWebAvatarUrl()
			}
			items = append(items, AccessGroupMemberItem{ID: member.Id, UserID: member.UserId, Username: username, AvatarURL: avatarURL, MemberRole: member.MemberRole, Status: member.Status})
		}
		result = append(result, ManagedAccessGroupItem{ID: group.Id, Name: group.Name, Applications: items})
	}
	return component.SuccessResponse(result)
}

func ReviewManagedAccessGroupApplication(req component.BetterRequest[struct {
	GroupID  uint64 `json:"groupId" validate:"required"`
	MemberID uint64 `json:"memberId" validate:"required"`
	Approve  bool   `json:"approve"`
}]) component.Response {
	groupIDs, err := accessGroupMembers.ManagedGroupIDsByUser(req.UserId)
	if err != nil || !lo.Contains(groupIDs, req.Params.GroupID) {
		return component.FailResponseCode(component.MessagePermissionDenied, nil)
	}
	if err := accessadminservice.ReviewApplication(req.Params.GroupID, req.Params.MemberID, req.Params.Approve); err != nil {
		return accessAdminFailure("review managed access group application", err)
	}
	return component.SuccessResponse(true)
}

type SaveCategoryAccessReq struct {
	CategoryID uint64                          `json:"categoryId" validate:"required"`
	Grants     []accessadminservice.GrantInput `json:"grants"`
	Strategy   string                          `json:"strategy" validate:"omitempty,oneof=keep_category remove_category"`
}

func SaveCategoryAccess(req component.BetterRequest[SaveCategoryAccessReq]) component.Response {
	if category.Get(req.Params.CategoryID).Id == 0 {
		return component.FailResponseCode(component.MessageAdminCategoryNotFound, nil)
	}
	if err := accessadminservice.ReplaceCategoryGrants(req.Params.CategoryID, req.Params.Grants, req.Params.Strategy); err != nil {
		var conflict accessadminservice.RestrictionConflictError
		if errors.As(err, &conflict) {
			return component.FailResponseCode(component.MessageOperationFailed, component.MessageParams{"restrictionConflictCount": conflict.Count})
		}
		return accessAdminFailure("save category access", err)
	}
	hotdataserve.ClearCategoryCache()
	hotdataserve.ClearTopicListCache()
	return component.SuccessResponse(true)
}

func PreviewCategoryRestriction(req component.BetterRequest[struct {
	CategoryID uint64 `json:"categoryId" validate:"required"`
}]) component.Response {
	count, err := accessadminservice.PreviewRestriction(req.Params.CategoryID)
	if err != nil {
		return accessAdminFailure("preview category restriction", err)
	}
	return component.SuccessResponse(map[string]int{"conflictCount": count})
}

func accessAdminFailure(operation string, err error) component.Response {
	slog.Error(operation+" failed", "err", err)
	if errors.Is(err, accessadminservice.ErrInvalidGroup) || errors.Is(err, accessadminservice.ErrSystemGroupImmutable) || errors.Is(err, accessadminservice.ErrInvalidMember) || errors.Is(err, accessadminservice.ErrInvalidGrant) || errors.Is(err, accessadminservice.ErrApplicationNotAllowed) || errors.Is(err, accesscontrol.ErrTooManyActiveGroups) {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	return component.FailResponseCode(component.MessageOperationFailed, nil)
}
