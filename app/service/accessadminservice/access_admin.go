package accessadminservice

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const MaxRestrictionConversionTopics = 10000

var (
	ErrInvalidGroup          = errors.New("invalid access group")
	ErrSystemGroupImmutable  = errors.New("system access group is immutable")
	ErrInvalidMember         = errors.New("invalid access group member")
	ErrInvalidGrant          = errors.New("invalid category grant")
	ErrRestrictionStrategy   = errors.New("category restriction conversion strategy is required")
	ErrRestrictionTooLarge   = errors.New("category restriction conversion is too large")
	ErrApplicationNotAllowed = errors.New("access group does not accept applications")
)

const (
	RestrictionKeepCategory   = "keep_category"
	RestrictionRemoveCategory = "remove_category"
)

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

type RestrictionConflictError struct {
	Count int
}

func (err RestrictionConflictError) Error() string {
	return fmt.Sprintf("%d multi-category topics require conversion", err.Count)
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

// testHookBeforeRestrictionTx lets a test interleave a concurrent write between
// the preview and the transaction, which is exactly the window this code has to
// be correct across. It is nil outside tests.
var testHookBeforeRestrictionTx func()

// rebuildTopicSearchDocument is a variable so tests can observe reindexing
// without a running Meilisearch.
var rebuildTopicSearchDocument = func(topic *topics.Entity) {
	firstPost := posts.Get(topic.FirstPostId)
	if _, err := searchservice.BuildSingleTopicSearchDocument(topic, &firstPost); err != nil {
		slog.Error("reindex converted topic search document", "topicId", topic.Id, "err", err)
	}
}

// PreviewRestriction sizes a pending conversion for the operator. It is not the
// decision: the conversion is decided again inside the transaction, on locked
// rows, because a topic can be created or edited into conflict in between.
func PreviewRestriction(categoryID uint64) (int, error) {
	var count int64
	err := restrictionConflictQuery(dbconnect.Connect(), categoryID).Count(&count).Error
	return int(count), err
}

func ReplaceCategoryGrants(categoryID uint64, grants []GrantInput, strategy string) error {
	canonical, groups, err := validateGrants(grants)
	if err != nil {
		return err
	}
	everyoneID := uint64(0)
	for _, group := range groups {
		if group.SystemKey != nil && *group.SystemKey == accessGroups.SystemKeyEveryone {
			everyoneID = group.Id
		}
	}
	if everyoneID == 0 {
		return ErrInvalidGrant
	}
	existing, err := categoryGroupPermissions.ByCategoryIDs([]uint64{categoryID})
	if err != nil {
		return err
	}
	wasPublic := enabledLevel(existing, everyoneID) >= categoryGroupPermissions.PermissionRead
	willBePublic := canonical[everyoneID] >= categoryGroupPermissions.PermissionRead
	restricting := wasPublic && !willBePublic
	if restricting {
		// Fail early and cheaply so the operator gets the strategy prompt or the
		// size error without opening a transaction.
		previewCount, err := PreviewRestriction(categoryID)
		if err != nil {
			return err
		}
		if previewCount > MaxRestrictionConversionTopics {
			return ErrRestrictionTooLarge
		}
		if previewCount > 0 && !validRestrictionStrategy(strategy) {
			return RestrictionConflictError{Count: previewCount}
		}
	}

	if testHookBeforeRestrictionTx != nil {
		testHookBeforeRestrictionTx()
	}

	var converted []topics.Entity
	err = dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		converted = nil
		var affected []topics.Entity
		if restricting {
			if err := restrictionConflictQuery(tx.Clauses(clause.Locking{Strength: "UPDATE"}), categoryID).Find(&affected).Error; err != nil {
				return err
			}
			// The preview above is advisory. What the conversion acts on is this
			// locked read, so a topic that became conflicting after the preview
			// is either converted here or aborts the whole change — it can never
			// be left behind as a [public, restricted] topic.
			if len(affected) > MaxRestrictionConversionTopics {
				return ErrRestrictionTooLarge
			}
			if len(affected) > 0 && !validRestrictionStrategy(strategy) {
				return RestrictionConflictError{Count: len(affected)}
			}
			for i := range affected {
				categoryIDs := convertedCategories(affected[i].CategoryIds, categoryID, strategy)
				if len(categoryIDs) == 0 {
					return ErrRestrictionStrategy
				}
				affected[i].CategoryIds = categoryIDs
				if err := tx.Save(&affected[i]).Error; err != nil {
					return err
				}
				if err := topicCategoryIndex.ReplaceTopicCategoriesWithDB(tx, affected[i].Id, categoryIDs); err != nil {
					return err
				}
			}
			converted = affected
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
	// The conversion rewrote topics.category_ids, so the Meilisearch documents
	// of those topics still carry the old category array and would keep them
	// findable under a category they are no longer in. Reindex after the
	// transaction commits: the search index cannot roll back with it.
	for i := range converted {
		rebuildTopicSearchDocument(&converted[i])
	}
	for groupID := range canonical {
		accesscontrol.InvalidateGroup(groupID)
	}
	for _, grant := range existing {
		accesscontrol.InvalidateGroup(grant.AccessGroupId)
	}
	return nil
}

func validRestrictionStrategy(strategy string) bool {
	return strategy == RestrictionKeepCategory || strategy == RestrictionRemoveCategory
}

func restrictionConflictQuery(db *gorm.DB, categoryID uint64) *gorm.DB {
	return db.Model(&topics.Entity{}).
		Where(`EXISTS (SELECT 1 FROM topic_category_index selected_idx WHERE selected_idx.topic_id = topics.id AND selected_idx.category_id = ? AND selected_idx.effective = 1)`, categoryID).
		Where(`(SELECT COUNT(*) FROM topic_category_index count_idx WHERE count_idx.topic_id = topics.id AND count_idx.effective = 1) > 1`)
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

func enabledLevel(grants []categoryGroupPermissions.Entity, groupID uint64) int8 {
	for _, grant := range grants {
		if grant.AccessGroupId == groupID && grant.Status == categoryGroupPermissions.StatusEnabled {
			return grant.PermissionLevel
		}
	}
	return 0
}

func convertedCategories(current []uint64, target uint64, strategy string) []uint64 {
	if strategy == RestrictionKeepCategory {
		return []uint64{target}
	}
	result := make([]uint64, 0, len(current))
	for _, categoryID := range current {
		if categoryID != 0 && categoryID != target {
			result = append(result, categoryID)
		}
	}
	return result
}

func validJoinMode(value string) bool {
	return value == accessGroups.JoinModeInviteOnly || value == accessGroups.JoinModeApplication
}
