package accessadminservice

import (
	"errors"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidGroup          = errors.New("invalid access group")
	ErrSystemGroupImmutable  = errors.New("system access group is immutable")
	ErrInvalidMember         = errors.New("invalid access group member")
	ErrInvalidGrant          = errors.New("invalid category grant")
	ErrApplicationNotAllowed = errors.New("access group does not accept applications")
)

type CategoryRestrictionConflictError struct {
	TopicCount int64
}

func (err *CategoryRestrictionConflictError) Error() string {
	return "category is used by multi-category topics"
}

type GroupInput struct {
	ID        uint64
	Name      string
	JoinMode  string
	Status    int8
	CreatedBy uint64
}

type GrantInput struct {
	AccessGroupID uint64 `json:"accessGroupId"`
	Level         int8   `json:"level"`
}

func SaveGroup(input GroupInput) (accessGroups.Entity, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 128 || !validJoinMode(input.JoinMode) || (input.Status != accessGroups.StatusEnabled && input.Status != accessGroups.StatusDisabled) {
		return accessGroups.Entity{}, ErrInvalidGroup
	}
	if input.ID == 0 {
		entity := accessGroups.Entity{Name: name, JoinMode: input.JoinMode, Status: input.Status, CreatedBy: input.CreatedBy}
		if err := accessGroups.Create(&entity); err != nil {
			return accessGroups.Entity{}, err
		}
		return entity, nil
	}
	entity, err := accessGroups.Get(input.ID)
	if err != nil || entity.Id == 0 {
		return accessGroups.Entity{}, ErrInvalidGroup
	}
	if entity.SystemKey != nil {
		return accessGroups.Entity{}, ErrSystemGroupImmutable
	}
	members, err := accessGroupMembers.ByGroupID(entity.Id)
	if err != nil {
		return accessGroups.Entity{}, err
	}
	entity.Name = name
	entity.JoinMode = input.JoinMode
	entity.Status = input.Status
	if err := accessGroups.Save(&entity); err != nil {
		return accessGroups.Entity{}, err
	}
	accesscontrol.InvalidateGroup(entity.Id)
	for _, member := range members {
		accesscontrol.InvalidateUser(member.UserId)
	}
	return entity, nil
}

func SaveCategoryWithDefaults(entity *category.Entity, create bool) error {
	if entity == nil {
		return errors.New("category is required")
	}
	invalidatedGroupIDs := make([]uint64, 0, 2)
	err := dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(entity).Error; err != nil {
			return err
		}
		if !create {
			return nil
		}
		var groups []accessGroups.Entity
		if err := tx.Where("system_key IN ? AND status = ?", []string{accessGroups.SystemKeyEveryone, accessGroups.SystemKeyRegistered}, accessGroups.StatusEnabled).Find(&groups).Error; err != nil {
			return err
		}
		levels := map[string]int8{accessGroups.SystemKeyEveryone: categoryGroupPermissions.PermissionRead, accessGroups.SystemKeyRegistered: categoryGroupPermissions.PermissionCreate}
		rows := make([]categoryGroupPermissions.Entity, 0, 2)
		for _, group := range groups {
			if group.SystemKey == nil || levels[*group.SystemKey] == 0 {
				continue
			}
			invalidatedGroupIDs = append(invalidatedGroupIDs, group.Id)
			rows = append(rows, categoryGroupPermissions.Entity{CategoryId: entity.Id, AccessGroupId: group.Id, PermissionLevel: levels[*group.SystemKey], Status: categoryGroupPermissions.StatusEnabled})
		}
		if len(rows) != 2 {
			return errors.New("required system access groups are missing")
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "category_id"}, {Name: "access_group_id"}},
			DoNothing: true,
		}).Create(&rows).Error
	})
	if err != nil {
		return err
	}
	for _, groupID := range invalidatedGroupIDs {
		accesscontrol.InvalidateGroup(groupID)
	}
	return nil
}

func DeleteCategory(entity *category.Entity) error {
	if entity == nil || entity.Id == 0 {
		return errors.New("category is required")
	}
	grants, err := categoryGroupPermissions.ByCategoryIDs([]uint64{entity.Id})
	if err != nil {
		return err
	}
	err = dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("category_id = ?", entity.Id).Delete(&categoryGroupPermissions.Entity{}).Error; err != nil {
			return err
		}
		return tx.Delete(entity).Error
	})
	if err != nil {
		return err
	}
	for _, grant := range grants {
		accesscontrol.InvalidateGroup(grant.AccessGroupId)
	}
	return nil
}

func DeleteGroup(groupID uint64) error {
	group, err := accessGroups.Get(groupID)
	if err != nil || group.Id == 0 {
		return ErrInvalidGroup
	}
	if group.SystemKey != nil {
		return ErrSystemGroupImmutable
	}
	var userIDs []uint64
	err = dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&accessGroupMembers.Entity{}).Where("access_group_id = ?", groupID).Pluck("user_id", &userIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("access_group_id = ?", groupID).Delete(&accessGroupMembers.Entity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("access_group_id = ?", groupID).Delete(&categoryGroupPermissions.Entity{}).Error; err != nil {
			return err
		}
		return tx.Delete(&group).Error
	})
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		accesscontrol.InvalidateUser(userID)
	}
	accesscontrol.InvalidateGroup(groupID)
	return nil
}

func SaveMember(groupID, userID uint64, role string, actorID uint64) (accessGroupMembers.Entity, error) {
	group, err := accessGroups.Get(groupID)
	if err != nil || group.Id == 0 || group.SystemKey != nil || group.Status != accessGroups.StatusEnabled || userID == 0 || (role != accessGroupMembers.MemberRoleMember && role != accessGroupMembers.MemberRoleManager) {
		return accessGroupMembers.Entity{}, ErrInvalidMember
	}
	entity, findErr := accessGroupMembers.GetByGroupUser(groupID, userID)
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return accessGroupMembers.Entity{}, findErr
	}
	if entity.Id == 0 || entity.Status != accessGroupMembers.StatusEnabled {
		count, err := accessGroupMembers.CountActiveCustomGroupsByUser(userID)
		if err != nil {
			return accessGroupMembers.Entity{}, err
		}
		if count >= accesscontrol.MaxActiveCustomGroups {
			return accessGroupMembers.Entity{}, accesscontrol.ErrTooManyActiveGroups
		}
	}
	entity.AccessGroupId = groupID
	entity.UserId = userID
	entity.MemberRole = role
	entity.Status = accessGroupMembers.StatusEnabled
	if entity.CreatedBy == 0 {
		entity.CreatedBy = actorID
	}
	if err := accessGroupMembers.Save(&entity); err != nil {
		return accessGroupMembers.Entity{}, err
	}
	accesscontrol.InvalidateUser(userID)
	return entity, nil
}

func DeleteMember(groupID, memberID uint64) error {
	members, err := accessGroupMembers.ByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.Id != memberID {
			continue
		}
		if err := accessGroupMembers.Delete(&member); err != nil {
			return err
		}
		accesscontrol.InvalidateUser(member.UserId)
		return nil
	}
	return ErrInvalidMember
}

func ApplyToGroup(groupID, userID uint64) error {
	group, err := accessGroups.Get(groupID)
	if err != nil || group.Id == 0 || group.SystemKey != nil || group.Status != accessGroups.StatusEnabled || group.JoinMode != accessGroups.JoinModeApplication || userID == 0 {
		return ErrApplicationNotAllowed
	}
	member, findErr := accessGroupMembers.GetByGroupUser(groupID, userID)
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return findErr
	}
	if member.Status == accessGroupMembers.StatusEnabled || member.Status == accessGroupMembers.StatusPending {
		return nil
	}
	member.AccessGroupId = groupID
	member.UserId = userID
	member.MemberRole = accessGroupMembers.MemberRoleMember
	member.Status = accessGroupMembers.StatusPending
	return accessGroupMembers.Save(&member)
}

func ReviewApplication(groupID, memberID uint64, approve bool) error {
	members, err := accessGroupMembers.ByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.Id != memberID || member.Status != accessGroupMembers.StatusPending {
			continue
		}
		if approve {
			count, err := accessGroupMembers.CountActiveCustomGroupsByUser(member.UserId)
			if err != nil {
				return err
			}
			if count >= accesscontrol.MaxActiveCustomGroups {
				return accesscontrol.ErrTooManyActiveGroups
			}
			member.Status = accessGroupMembers.StatusEnabled
		} else {
			member.Status = accessGroupMembers.StatusDisabled
		}
		if err := accessGroupMembers.Save(&member); err != nil {
			return err
		}
		accesscontrol.InvalidateUser(member.UserId)
		return nil
	}
	return ErrInvalidMember
}

// ReplaceCategoryGrants swaps a category's grants. A public category cannot be
// made restricted while it is still attached to multi-category topics: the
// single-restricted-category invariant must hold before visibility narrows.
func ReplaceCategoryGrants(categoryID uint64, grants []GrantInput) error {
	canonical, _, err := validateGrants(grants)
	if err != nil {
		return err
	}
	existing, err := categoryGroupPermissions.ByCategoryIDs([]uint64{categoryID})
	if err != nil {
		return err
	}

	err = dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		everyoneID, publicCategories, err := accesscontrol.LockPublicCategoryStateWithDB(tx, []uint64{categoryID})
		if err != nil {
			return err
		}
		_, currentlyPublic := publicCategories[categoryID]
		willBePublic := canonical[everyoneID] >= categoryGroupPermissions.PermissionRead
		if currentlyPublic && !willBePublic {
			counts, err := topicCategoryIndex.MultiCategoryTopicCountsWithDB(tx, []uint64{categoryID})
			if err != nil {
				return err
			}
			if counts[categoryID] > 0 {
				return &CategoryRestrictionConflictError{TopicCount: counts[categoryID]}
			}
		}
		if err := tx.Model(&categoryGroupPermissions.Entity{}).Where("category_id = ?", categoryID).
			Updates(map[string]any{"status": categoryGroupPermissions.StatusDisabled, "permission_level": 0}).Error; err != nil {
			return err
		}
		rows := make([]categoryGroupPermissions.Entity, 0, len(canonical))
		for groupID, level := range canonical {
			if level == 0 {
				continue
			}
			rows = append(rows, categoryGroupPermissions.Entity{CategoryId: categoryID, AccessGroupId: groupID, PermissionLevel: level, Status: categoryGroupPermissions.StatusEnabled})
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "category_id"}, {Name: "access_group_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"permission_level", "status", "updated_at"}),
		}).Create(&rows).Error
	})
	if err != nil {
		return err
	}
	for groupID := range canonical {
		accesscontrol.InvalidateGroup(groupID)
	}
	for _, grant := range existing {
		accesscontrol.InvalidateGroup(grant.AccessGroupId)
	}
	return nil
}

func validateGrants(grants []GrantInput) (map[uint64]int8, map[uint64]accessGroups.Entity, error) {
	all, err := accessGroups.All()
	if err != nil {
		return nil, nil, err
	}
	groups := make(map[uint64]accessGroups.Entity, len(all))
	for _, group := range all {
		if group.Status == accessGroups.StatusEnabled {
			groups[group.Id] = group
		}
	}
	canonical := make(map[uint64]int8, len(grants))
	for _, grant := range grants {
		if groups[grant.AccessGroupID].Id == 0 || grant.Level < 0 || grant.Level > categoryGroupPermissions.PermissionManage {
			return nil, nil, ErrInvalidGrant
		}
		canonical[grant.AccessGroupID] = grant.Level
	}
	return canonical, groups, nil
}

func validJoinMode(value string) bool {
	return value == accessGroups.JoinModeInviteOnly || value == accessGroups.JoinModeApplication
}
