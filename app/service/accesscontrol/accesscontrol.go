package accesscontrol

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/cacheconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/service/moderationservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

const (
	MaxActiveCustomGroups = 32
	metadataTTL           = 2 * time.Minute
)

var ErrTooManyActiveGroups = errors.New("user exceeds active access group limit")

var (
	ErrCategoryRequired               = errors.New("at least one category is required")
	ErrTooManyCategories              = errors.New("at most three categories are allowed")
	ErrCategoryPermissionDenied       = errors.New("category capability is insufficient")
	ErrRestrictedCategoryMustBeSingle = errors.New("a restricted topic must use exactly one category")
)

type TopicCategoryWrite struct {
	Current    []uint64
	Next       []uint64
	Publishing bool
	NewTopic   bool
}

type Capability int8

const (
	CapabilityNone   Capability = 0
	CapabilityRead   Capability = Capability(categoryGroupPermissions.PermissionRead)
	CapabilityReply  Capability = Capability(categoryGroupPermissions.PermissionReply)
	CapabilityCreate Capability = Capability(categoryGroupPermissions.PermissionCreate)
	CapabilityManage Capability = Capability(categoryGroupPermissions.PermissionManage)
)

type CategoryGrant struct {
	CategoryID uint64
	Capability Capability
}

type Store interface {
	SystemGroupIDs() (map[string]uint64, error)
	ActiveCustomGroupIDs(userID uint64) ([]uint64, error)
	EnabledCategoryGrants(groupID uint64) ([]CategoryGrant, error)
}

type Snapshot struct {
	levels            map[uint64]Capability
	globalManage      bool
	listCacheAudience string
	listCacheable     bool
}

func (snapshot Snapshot) Capability(categoryID uint64) Capability {
	if snapshot.globalManage {
		return CapabilityManage
	}
	return snapshot.levels[categoryID]
}

func (snapshot Snapshot) CanReadCategory(categoryID uint64) bool {
	return snapshot.Capability(categoryID) >= CapabilityRead
}

func (snapshot Snapshot) CanReplyCategory(categoryID uint64) bool {
	return snapshot.Capability(categoryID) >= CapabilityReply
}

func (snapshot Snapshot) CanCreateCategory(categoryID uint64) bool {
	return snapshot.Capability(categoryID) >= CapabilityCreate
}

func (snapshot Snapshot) CanManageCategory(categoryID uint64) bool {
	return snapshot.Capability(categoryID) >= CapabilityManage
}

func (snapshot Snapshot) CanReadAllCategories(categoryIDs []uint64) bool {
	return snapshot.canAllCategories(categoryIDs, CapabilityRead)
}

func (snapshot Snapshot) CanReplyAllCategories(categoryIDs []uint64) bool {
	return snapshot.canAllCategories(categoryIDs, CapabilityReply)
}

func (snapshot Snapshot) CanCreateAllCategories(categoryIDs []uint64) bool {
	return snapshot.canAllCategories(categoryIDs, CapabilityCreate)
}

func (snapshot Snapshot) CanManageAllCategories(categoryIDs []uint64) bool {
	return snapshot.canAllCategories(categoryIDs, CapabilityManage)
}

func ValidateCategorySelection(actor Snapshot, everyone Snapshot, categoryIDs []uint64, required Capability) ([]uint64, error) {
	categoryIDs = uniqueNonZeroPreservingOrder(categoryIDs)
	if len(categoryIDs) == 0 {
		return nil, ErrCategoryRequired
	}
	if len(categoryIDs) > 3 {
		return nil, ErrTooManyCategories
	}
	for _, categoryID := range categoryIDs {
		if actor.Capability(categoryID) < required {
			return nil, ErrCategoryPermissionDenied
		}
	}
	if SelectionIsRestricted(everyone, categoryIDs) && len(categoryIDs) != 1 {
		return nil, ErrRestrictedCategoryMustBeSingle
	}
	return categoryIDs, nil
}

// ValidateTopicCategoryWrite centralizes category authorization for topic
// creation, editing, publishing, and category changes. Moving a topic into or
// out of a restricted category is deliberately a management operation.
func ValidateTopicCategoryWrite(actor Snapshot, everyone Snapshot, input TopicCategoryWrite) ([]uint64, error) {
	if input.NewTopic {
		return ValidateCategorySelection(actor, everyone, input.Next, CapabilityCreate)
	}
	current := uniqueNonZeroPreservingOrder(input.Current)
	if len(current) == 0 || !actor.CanReadAllCategories(current) {
		return nil, ErrCategoryPermissionDenied
	}
	required := CapabilityRead
	changed := !SameCategorySelection(current, input.Next)
	if input.Publishing || changed {
		required = CapabilityCreate
	}
	next, err := ValidateCategorySelection(actor, everyone, input.Next, required)
	if err != nil {
		return nil, err
	}
	if changed && (SelectionIsRestricted(everyone, current) || SelectionIsRestricted(everyone, next)) {
		union := append(append([]uint64(nil), current...), next...)
		if !actor.CanManageAllCategories(union) {
			return nil, ErrCategoryPermissionDenied
		}
	}
	return next, nil
}

func SelectionIsRestricted(everyone Snapshot, categoryIDs []uint64) bool {
	for _, categoryID := range uniqueNonZeroPreservingOrder(categoryIDs) {
		if !everyone.CanReadCategory(categoryID) {
			return true
		}
	}
	return false
}

func SameCategorySelection(a []uint64, b []uint64) bool {
	a = uniqueNonZero(a)
	b = uniqueNonZero(b)
	return slices.Equal(a, b)
}

func (snapshot Snapshot) ReadableCategoryIDs() []uint64 {
	return snapshot.categoryIDsAtLeast(CapabilityRead)
}

func (snapshot Snapshot) CreatableCategoryIDs() []uint64 {
	return snapshot.categoryIDsAtLeast(CapabilityCreate)
}

func (snapshot Snapshot) HasGlobalManage() bool {
	return snapshot.globalManage
}

func (snapshot Snapshot) ListCacheAudience() (string, bool) {
	return snapshot.listCacheAudience, snapshot.listCacheable && snapshot.listCacheAudience != ""
}

func (snapshot Snapshot) categoryIDsAtLeast(required Capability) []uint64 {
	categoryIDs := make([]uint64, 0, len(snapshot.levels))
	for categoryID, level := range snapshot.levels {
		if categoryID > 0 && level >= required {
			categoryIDs = append(categoryIDs, categoryID)
		}
	}
	slices.Sort(categoryIDs)
	return categoryIDs
}

func (snapshot Snapshot) canAllCategories(categoryIDs []uint64, required Capability) bool {
	categoryIDs = uniqueNonZero(categoryIDs)
	if len(categoryIDs) == 0 {
		return false
	}
	for _, categoryID := range categoryIDs {
		if snapshot.Capability(categoryID) < required {
			return false
		}
	}
	return true
}

type Resolver struct {
	store Store

	membershipCache *localcache.Cache[[]uint64]
	grantCache      *localcache.Cache[[]CategoryGrant]
	systemCache     *localcache.Cache[map[string]uint64]

	globalContentManager func(userID uint64) bool
	moderationScope      func(userID uint64) (global bool, categoryIDs []uint64)
}

func NewResolver(
	store Store,
	globalContentManager func(userID uint64) bool,
	moderationScope func(userID uint64) (bool, []uint64),
) *Resolver {
	return &Resolver{
		store:                store,
		membershipCache:      &localcache.Cache[[]uint64]{MaxEntries: cacheconfig.Current().AccessGroupMembers},
		grantCache:           &localcache.Cache[[]CategoryGrant]{MaxEntries: cacheconfig.Current().AccessGroupGrants},
		systemCache:          &localcache.Cache[map[string]uint64]{MaxEntries: 1},
		globalContentManager: globalContentManager,
		moderationScope:      moderationScope,
	}
}

func (resolver *Resolver) Resolve(userID uint64) (Snapshot, error) {
	if resolver == nil || resolver.store == nil {
		return Snapshot{}, errors.New("access control resolver is not configured")
	}
	systemGroupIDs, err := resolver.systemCache.GetOrLoadE("system", resolver.store.SystemGroupIDs, metadataTTL)
	if err != nil {
		return Snapshot{}, fmt.Errorf("load system access groups: %w", err)
	}
	everyoneID := systemGroupIDs[accessGroups.SystemKeyEveryone]
	registeredID := systemGroupIDs[accessGroups.SystemKeyRegistered]
	if everyoneID == 0 || registeredID == 0 {
		return Snapshot{}, errors.New("required system access groups are missing")
	}

	groupIDs := []uint64{everyoneID}
	listCacheAudience := "guest"
	listCacheable := true
	if userID != 0 {
		listCacheAudience = "registered"
		groupIDs = append(groupIDs, registeredID)
		customGroupIDs, err := resolver.membershipCache.GetOrLoadE(
			strconv.FormatUint(userID, 10),
			func() ([]uint64, error) { return resolver.store.ActiveCustomGroupIDs(userID) },
			metadataTTL,
		)
		if err != nil {
			return Snapshot{}, fmt.Errorf("load access group memberships: %w", err)
		}
		customGroupIDs = uniqueNonZero(customGroupIDs)
		if len(customGroupIDs) > MaxActiveCustomGroups {
			return Snapshot{}, fmt.Errorf("%w: user %d has %d", ErrTooManyActiveGroups, userID, len(customGroupIDs))
		}
		if len(customGroupIDs) > 0 {
			listCacheable = false
		}
		groupIDs = append(groupIDs, customGroupIDs...)
	}
	groupIDs = uniqueNonZero(groupIDs)

	levels := make(map[uint64]Capability)
	for _, groupID := range groupIDs {
		grants, err := resolver.grantCache.GetOrLoadE(
			strconv.FormatUint(groupID, 10),
			func() ([]CategoryGrant, error) { return resolver.store.EnabledCategoryGrants(groupID) },
			metadataTTL,
		)
		if err != nil {
			return Snapshot{}, fmt.Errorf("load category grants for group %d: %w", groupID, err)
		}
		mergeGrants(levels, grants)
	}

	snapshot := Snapshot{
		levels:            levels,
		listCacheAudience: listCacheAudience,
		listCacheable:     listCacheable,
	}
	if userID == 0 {
		return snapshot, nil
	}
	if resolver.globalContentManager != nil && resolver.globalContentManager(userID) {
		snapshot.globalManage = true
		snapshot.listCacheable = false
		return snapshot, nil
	}
	if resolver.moderationScope != nil {
		global, categoryIDs := resolver.moderationScope(userID)
		if global {
			snapshot.globalManage = true
			snapshot.listCacheable = false
			return snapshot, nil
		}
		for _, categoryID := range categoryIDs {
			if categoryID > 0 {
				snapshot.levels[categoryID] = CapabilityManage
				snapshot.listCacheable = false
			}
		}
	}
	return snapshot, nil
}

func (resolver *Resolver) InvalidateUser(userID uint64) {
	if resolver != nil && userID != 0 {
		resolver.membershipCache.Delete(strconv.FormatUint(userID, 10))
	}
}

func (resolver *Resolver) InvalidateGroup(groupID uint64) {
	if resolver != nil && groupID != 0 {
		resolver.grantCache.Delete(strconv.FormatUint(groupID, 10))
	}
}

func (resolver *Resolver) InvalidateSystemGroups() {
	if resolver != nil {
		resolver.systemCache.Clear()
	}
}

func mergeGrants(levels map[uint64]Capability, grants []CategoryGrant) {
	for _, grant := range grants {
		if grant.CategoryID == 0 || grant.Capability < CapabilityRead || grant.Capability > CapabilityManage {
			continue
		}
		if levels[grant.CategoryID] < grant.Capability {
			levels[grant.CategoryID] = grant.Capability
		}
	}
}

func uniqueNonZero(values []uint64) []uint64 {
	result := uniqueNonZeroPreservingOrder(values)
	slices.Sort(result)
	return result
}

func uniqueNonZeroPreservingOrder(values []uint64) []uint64 {
	result := make([]uint64, 0, len(values))
	seen := make(map[uint64]struct{}, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

type modelStore struct{}

func (modelStore) SystemGroupIDs() (map[string]uint64, error) {
	groups, err := accessGroups.GetBySystemKeys([]string{
		accessGroups.SystemKeyEveryone,
		accessGroups.SystemKeyRegistered,
	})
	if err != nil {
		return nil, err
	}
	result := make(map[string]uint64, len(groups))
	for _, group := range groups {
		if group.SystemKey != nil {
			result[*group.SystemKey] = group.Id
		}
	}
	return result, nil
}

func (modelStore) ActiveCustomGroupIDs(userID uint64) ([]uint64, error) {
	return accessGroupMembers.ActiveGroupIDsByUser(userID)
}

func (modelStore) EnabledCategoryGrants(groupID uint64) ([]CategoryGrant, error) {
	rows, err := categoryGroupPermissions.EnabledByGroupID(groupID)
	if err != nil {
		return nil, err
	}
	result := make([]CategoryGrant, 0, len(rows))
	for _, row := range rows {
		result = append(result, CategoryGrant{
			CategoryID: row.CategoryId,
			Capability: Capability(row.PermissionLevel),
		})
	}
	return result, nil
}

func defaultGlobalContentManager(userID uint64) bool {
	roleID, ok := userservice.GetUserRoleId(userID)
	return ok && (permission.CheckRole(roleID, permission.Admin) || permission.CheckRole(roleID, permission.TopicsManager))
}

var Default = NewResolver(modelStore{}, defaultGlobalContentManager, moderationservice.ScopeForUser)

func Resolve(userID uint64) (Snapshot, error) {
	return Default.Resolve(userID)
}

func InvalidateUser(userID uint64) {
	Default.InvalidateUser(userID)
}

func InvalidateGroup(groupID uint64) {
	Default.InvalidateGroup(groupID)
}

func InvalidateSystemGroups() {
	Default.InvalidateSystemGroups()
}
