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
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
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
	ErrCategoryRequired         = errors.New("at least one category is required")
	ErrTooManyCategories        = errors.New("at most three categories are allowed")
	ErrCategoryPermissionDenied = errors.New("category capability is insufficient")
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

type batchReadableStore interface {
	ActiveUserIDsWithCategoryCapability(userIDs []uint64, categoryID uint64, required Capability) ([]uint64, error)
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

// MainCategoryOf returns the category a topic draws its visibility from: the
// first one selected. The rest are auxiliary tags and never widen or narrow
// who can read the topic.
func MainCategoryOf(categoryIDs []uint64) uint64 {
	for _, categoryID := range categoryIDs {
		if categoryID != 0 {
			return categoryID
		}
	}
	return 0
}

func ValidateCategorySelection(actor Snapshot, categoryIDs []uint64, required Capability) ([]uint64, error) {
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
	return categoryIDs, nil
}

// ValidateTopicCategoryWrite centralizes category authorization for topic
// creation, editing, publishing, and category changes. Only the main category
// decides who can read a topic, so only a change of main category changes
// visibility, and only that is escalated to a management operation.
func ValidateTopicCategoryWrite(actor Snapshot, everyone Snapshot, input TopicCategoryWrite) ([]uint64, error) {
	if input.NewTopic {
		return ValidateCategorySelection(actor, input.Next, CapabilityCreate)
	}
	current := uniqueNonZeroPreservingOrder(input.Current)
	currentMain := MainCategoryOf(current)
	if currentMain == 0 || !actor.CanReadCategory(currentMain) {
		return nil, ErrCategoryPermissionDenied
	}
	required := CapabilityRead
	changed := !SameCategorySelection(current, input.Next)
	if input.Publishing || changed {
		required = CapabilityCreate
	}
	next, err := ValidateCategorySelection(actor, input.Next, required)
	if err != nil {
		return nil, err
	}
	nextMain := MainCategoryOf(next)
	if currentMain != nextMain && (!everyone.CanReadCategory(currentMain) || !everyone.CanReadCategory(nextMain)) {
		if !actor.CanManageCategory(currentMain) || !actor.CanManageCategory(nextMain) {
			return nil, ErrCategoryPermissionDenied
		}
	}
	return next, nil
}

// SameCategorySelection compares order-sensitively: the first entry is the main
// category, so reordering the same three categories is a real change.
func SameCategorySelection(a []uint64, b []uint64) bool {
	return slices.Equal(uniqueNonZeroPreservingOrder(a), uniqueNonZeroPreservingOrder(b))
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

type Resolver struct {
	store Store

	membershipCache *localcache.Cache[[]uint64]
	grantCache      *localcache.Cache[[]CategoryGrant]
	systemCache     *localcache.Cache[map[string]uint64]

	globalContentManager func(userID uint64) bool
	moderationScope      func(userID uint64) (global bool, categoryIDs []uint64)
	batchContentManagers func(userIDs []uint64) []uint64
	batchModerators      func(userIDs []uint64, categoryID uint64) []uint64
}

// FilterReadableUserIDs filters a batch for a single category while preserving
// input order. The default resolver uses one membership query for the entire
// batch; custom resolvers fall back to Resolve so tests and integrations retain
// exactly the same semantics without implementing the optional batch store.
func (resolver *Resolver) FilterReadableUserIDs(userIDs []uint64, categoryID uint64) ([]uint64, error) {
	userIDs = uniqueNonZeroPreservingOrder(userIDs)
	if len(userIDs) == 0 || categoryID == 0 {
		return []uint64{}, nil
	}
	if resolver == nil || resolver.store == nil {
		return nil, errors.New("access control resolver is not configured")
	}
	batchStore, supportsBatch := resolver.store.(batchReadableStore)
	if !supportsBatch {
		return resolver.filterReadableUserIDsIndividually(userIDs, categoryID)
	}

	systemGroupIDs, err := resolver.loadSystemGroupIDs()
	if err != nil {
		return nil, err
	}
	for _, systemKey := range []string{accessGroups.SystemKeyEveryone, accessGroups.SystemKeyRegistered} {
		groupID := systemGroupIDs[systemKey]
		if groupID == 0 {
			return nil, errors.New("required system access groups are missing")
		}
		grants, err := resolver.loadGroupGrants(groupID)
		if err != nil {
			return nil, err
		}
		for _, grant := range grants {
			if grant.CategoryID == categoryID && grant.Capability >= CapabilityRead {
				return userIDs, nil
			}
		}
	}

	readableUserIDs, err := batchStore.ActiveUserIDsWithCategoryCapability(userIDs, categoryID, CapabilityRead)
	if err != nil {
		return nil, fmt.Errorf("filter access group members: %w", err)
	}
	readable := make(map[uint64]struct{}, len(readableUserIDs))
	for _, userID := range readableUserIDs {
		readable[userID] = struct{}{}
	}
	if resolver.batchContentManagers != nil {
		for _, userID := range resolver.batchContentManagers(userIDs) {
			readable[userID] = struct{}{}
		}
	}
	if resolver.batchModerators != nil {
		for _, userID := range resolver.batchModerators(userIDs, categoryID) {
			readable[userID] = struct{}{}
		}
	}
	if resolver.batchContentManagers == nil || resolver.batchModerators == nil {
		for _, userID := range userIDs {
			if _, ok := readable[userID]; ok {
				continue
			}
			if resolver.batchContentManagers == nil && resolver.globalContentManager != nil && resolver.globalContentManager(userID) {
				readable[userID] = struct{}{}
				continue
			}
			if resolver.batchModerators == nil && resolver.moderationScope != nil {
				global, categoryIDs := resolver.moderationScope(userID)
				if global || slices.Contains(categoryIDs, categoryID) {
					readable[userID] = struct{}{}
				}
			}
		}
	}

	result := make([]uint64, 0, len(readable))
	for _, userID := range userIDs {
		if _, ok := readable[userID]; ok {
			result = append(result, userID)
		}
	}
	return result, nil
}

func (resolver *Resolver) filterReadableUserIDsIndividually(userIDs []uint64, categoryID uint64) ([]uint64, error) {
	result := make([]uint64, 0, len(userIDs))
	for _, userID := range userIDs {
		snapshot, err := resolver.Resolve(userID)
		if err != nil {
			return nil, err
		}
		if snapshot.CanReadCategory(categoryID) {
			result = append(result, userID)
		}
	}
	return result, nil
}

func (resolver *Resolver) loadSystemGroupIDs() (map[string]uint64, error) {
	groupIDs, err := resolver.systemCache.GetOrLoadE("system", resolver.store.SystemGroupIDs, metadataTTL)
	if err != nil {
		return nil, fmt.Errorf("load system access groups: %w", err)
	}
	return groupIDs, nil
}

func (resolver *Resolver) loadGroupGrants(groupID uint64) ([]CategoryGrant, error) {
	grants, err := resolver.grantCache.GetOrLoadE(
		strconv.FormatUint(groupID, 10),
		func() ([]CategoryGrant, error) { return resolver.store.EnabledCategoryGrants(groupID) },
		metadataTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("load category grants for group %d: %w", groupID, err)
	}
	return grants, nil
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
	systemGroupIDs, err := resolver.loadSystemGroupIDs()
	if err != nil {
		return Snapshot{}, err
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
		grants, err := resolver.loadGroupGrants(groupID)
		if err != nil {
			return Snapshot{}, err
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

func (modelStore) ActiveUserIDsWithCategoryCapability(userIDs []uint64, categoryID uint64, required Capability) ([]uint64, error) {
	return accessGroupMembers.ActiveUserIDsWithCategoryCapability(userIDs, categoryID, int8(required))
}

func defaultGlobalContentManager(userID uint64) bool {
	roleID, ok := userservice.GetUserRoleId(userID)
	return ok && (permission.CheckRole(roleID, permission.Admin) || permission.CheckRole(roleID, permission.TopicsManager))
}

func defaultBatchContentManagers(userIDs []uint64) []uint64 {
	userMap := users.GetMapByIds(userIDs)
	roleIDs := make([]uint64, 0, len(userMap))
	for _, user := range userMap {
		if user != nil && user.RoleId > 0 {
			roleIDs = append(roleIDs, user.RoleId)
		}
	}
	roleIDs = uniqueNonZero(roleIDs)
	if len(roleIDs) == 0 {
		return nil
	}
	permissionsByRole := rolePermissionRs.GetRsGroupByRoleIds(roleIDs)
	result := make([]uint64, 0)
	for _, userID := range userIDs {
		user := userMap[userID]
		if user == nil {
			continue
		}
		permissions := permissionsByRole[user.RoleId]
		if slices.Contains(permissions, permission.Admin.Id()) || slices.Contains(permissions, permission.TopicsManager.Id()) {
			result = append(result, userID)
		}
	}
	return result
}

func newDefaultResolver() *Resolver {
	resolver := NewResolver(modelStore{}, defaultGlobalContentManager, moderationservice.ScopeForUser)
	resolver.batchContentManagers = defaultBatchContentManagers
	resolver.batchModerators = moderationservice.FilterUserIDsForCategory
	return resolver
}

var Default = newDefaultResolver()

func Resolve(userID uint64) (Snapshot, error) {
	return Default.Resolve(userID)
}

func FilterReadableUserIDs(userIDs []uint64, categoryID uint64) ([]uint64, error) {
	return Default.FilterReadableUserIDs(userIDs, categoryID)
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
