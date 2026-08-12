package accesscontrol

import (
	"errors"
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
)

type fakeStore struct {
	system      map[string]uint64
	memberships map[uint64][]uint64
	grants      map[uint64][]CategoryGrant
	err         error
	memberLoads int
	grantLoads  map[uint64]int
}

func (store *fakeStore) SystemGroupIDs() (map[string]uint64, error) {
	if store.err != nil {
		return nil, store.err
	}
	return store.system, nil
}

func (store *fakeStore) ActiveCustomGroupIDs(userID uint64) ([]uint64, error) {
	store.memberLoads++
	return append([]uint64(nil), store.memberships[userID]...), nil
}

func (store *fakeStore) EnabledCategoryGrants(groupID uint64) ([]CategoryGrant, error) {
	store.grantLoads[groupID]++
	return append([]CategoryGrant(nil), store.grants[groupID]...), nil
}

func TestResolveMergesSystemCustomAndModeratorGrants(t *testing.T) {
	store := newFakeStore()
	store.memberships[7] = []uint64{30, 20, 30}
	store.grants[10] = []CategoryGrant{{CategoryID: 1, Capability: CapabilityRead}}
	store.grants[11] = []CategoryGrant{{CategoryID: 1, Capability: CapabilityCreate}, {CategoryID: 2, Capability: CapabilityCreate}}
	store.grants[20] = []CategoryGrant{{CategoryID: 3, Capability: CapabilityRead}}
	store.grants[30] = []CategoryGrant{{CategoryID: 3, Capability: CapabilityReply}, {CategoryID: 4, Capability: CapabilityCreate}}

	resolver := NewResolver(store, func(uint64) bool { return false }, func(userID uint64) (bool, []uint64) {
		if userID == 7 {
			return false, []uint64{5}
		}
		return false, nil
	})
	snapshot, err := resolver.Resolve(7)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	assertCapability(t, snapshot, 1, CapabilityCreate)
	assertCapability(t, snapshot, 2, CapabilityCreate)
	assertCapability(t, snapshot, 3, CapabilityReply)
	assertCapability(t, snapshot, 4, CapabilityCreate)
	assertCapability(t, snapshot, 5, CapabilityManage)
	if got := snapshot.ReadableCategoryIDs(); !reflect.DeepEqual(got, []uint64{1, 2, 3, 4, 5}) {
		t.Fatalf("ReadableCategoryIDs = %v", got)
	}
	if _, cacheable := snapshot.ListCacheAudience(); cacheable {
		t.Fatal("custom-group audience must not cache list results")
	}
}

func TestGuestOnlyReceivesEveryoneGrants(t *testing.T) {
	store := newFakeStore()
	store.grants[10] = []CategoryGrant{{CategoryID: 1, Capability: CapabilityRead}}
	store.grants[11] = []CategoryGrant{{CategoryID: 2, Capability: CapabilityCreate}}
	resolver := NewResolver(store, nil, nil)

	snapshot, err := resolver.Resolve(0)
	if err != nil {
		t.Fatalf("Resolve guest: %v", err)
	}
	if !snapshot.CanReadCategory(1) || snapshot.CanReadCategory(2) {
		t.Fatalf("guest snapshot has unexpected capabilities: %+v", snapshot.ReadableCategoryIDs())
	}
	if store.memberLoads != 0 {
		t.Fatalf("guest membership loads = %d, want 0", store.memberLoads)
	}
	if audience, cacheable := snapshot.ListCacheAudience(); audience != "guest" || !cacheable {
		t.Fatalf("guest list cache audience = %q, %v", audience, cacheable)
	}
}

func TestMainCategoryAloneDecidesReadability(t *testing.T) {
	snapshot := Snapshot{levels: map[uint64]Capability{
		1: CapabilityCreate,
		2: CapabilityRead,
	}}
	if !snapshot.CanReadCategory(MainCategoryOf([]uint64{2, 1})) {
		t.Fatal("a readable main category should be readable")
	}
	if snapshot.CanReadCategory(MainCategoryOf([]uint64{3, 1, 2})) {
		t.Fatal("an unreadable main category must fail closed whatever the auxiliary tags are")
	}
	if MainCategoryOf(nil) != 0 || MainCategoryOf([]uint64{0, 5}) != 5 {
		t.Fatal("main category must be the first non-zero entry")
	}
	if snapshot.CanReadCategory(MainCategoryOf(nil)) {
		t.Fatal("a topic with no categories must fail closed")
	}
}

func TestValidateCategorySelectionPreservesOrder(t *testing.T) {
	actor := Snapshot{levels: map[uint64]Capability{1: CapabilityCreate, 2: CapabilityCreate, 3: CapabilityCreate}}
	everyone := Snapshot{levels: map[uint64]Capability{1: CapabilityRead, 2: CapabilityRead}}

	got, err := ValidateCategorySelection(actor, []uint64{2, 1, 2}, CapabilityCreate)
	if err != nil || !reflect.DeepEqual(got, []uint64{2, 1}) {
		t.Fatalf("public selection = %v, %v", got, err)
	}
	// 3 is restricted. Mixing it with a public category is allowed now: only the
	// first entry decides visibility, so there is no combination to forbid.
	got, err = ValidateCategorySelection(actor, []uint64{3, 1}, CapabilityCreate)
	if err != nil || !reflect.DeepEqual(got, []uint64{3, 1}) {
		t.Fatalf("restricted-main selection = %v, %v", got, err)
	}
	if _, err := ValidateCategorySelection(everyone, []uint64{3}, CapabilityCreate); !errors.Is(err, ErrCategoryPermissionDenied) {
		t.Fatalf("insufficient selection error = %v", err)
	}
}

func TestValidateTopicCategoryWritePolicies(t *testing.T) {
	everyone := Snapshot{levels: map[uint64]Capability{1: CapabilityRead, 2: CapabilityRead}}
	creator := Snapshot{levels: map[uint64]Capability{1: CapabilityCreate, 2: CapabilityCreate, 3: CapabilityCreate}}
	manager := Snapshot{levels: map[uint64]Capability{1: CapabilityManage, 3: CapabilityManage}}

	if got, err := ValidateTopicCategoryWrite(creator, everyone, TopicCategoryWrite{Next: []uint64{2, 1}, NewTopic: true}); err != nil || !reflect.DeepEqual(got, []uint64{2, 1}) {
		t.Fatalf("new topic categories = %v, %v", got, err)
	}
	if _, err := ValidateTopicCategoryWrite(everyone, everyone, TopicCategoryWrite{Current: []uint64{1}, Next: []uint64{1}, Publishing: true}); !errors.Is(err, ErrCategoryPermissionDenied) {
		t.Fatalf("publish without create error = %v", err)
	}
	if _, err := ValidateTopicCategoryWrite(creator, everyone, TopicCategoryWrite{Current: []uint64{1}, Next: []uint64{3}}); !errors.Is(err, ErrCategoryPermissionDenied) {
		t.Fatalf("restricted move without manage error = %v", err)
	}
	if got, err := ValidateTopicCategoryWrite(manager, everyone, TopicCategoryWrite{Current: []uint64{1}, Next: []uint64{3}}); err != nil || !reflect.DeepEqual(got, []uint64{3}) {
		t.Fatalf("managed restricted move = %v, %v", got, err)
	}
	// Auxiliary categories carry no visibility, so tagging a restricted category
	// onto a public topic is an ordinary edit, not a management operation.
	if got, err := ValidateTopicCategoryWrite(creator, everyone, TopicCategoryWrite{Current: []uint64{1}, Next: []uint64{1, 3}}); err != nil || !reflect.DeepEqual(got, []uint64{1, 3}) {
		t.Fatalf("adding a restricted auxiliary category = %v, %v", got, err)
	}
	// Reordering the same categories promotes 3 to main: that is a visibility
	// change and needs manage, which is why selection comparison is ordered.
	if _, err := ValidateTopicCategoryWrite(creator, everyone, TopicCategoryWrite{Current: []uint64{1, 3}, Next: []uint64{3, 1}}); !errors.Is(err, ErrCategoryPermissionDenied) {
		t.Fatalf("promoting a restricted category to main error = %v", err)
	}
}

// Removing the main category promotes the next one, which changes who can read
// the topic. That is why the promotion goes through the same manage gate as any
// other main-category change, and why an author cannot perform it.
func TestRemovingMainCategoryPromotesTheNextOneUnderTheManageGate(t *testing.T) {
	everyone := Snapshot{levels: map[uint64]Capability{1: CapabilityRead, 2: CapabilityRead}}
	creator := Snapshot{levels: map[uint64]Capability{1: CapabilityCreate, 2: CapabilityCreate, 3: CapabilityCreate}}
	manager := Snapshot{levels: map[uint64]Capability{1: CapabilityManage, 2: CapabilityManage, 3: CapabilityManage}}

	// Both public: the promotion does not change the audience, so it is an
	// ordinary edit and the author may do it.
	got, err := ValidateTopicCategoryWrite(creator, everyone, TopicCategoryWrite{Current: []uint64{1, 2}, Next: []uint64{2}})
	if err != nil || MainCategoryOf(got) != 2 {
		t.Fatalf("public promotion = %v, %v, want main 2", got, err)
	}
	// Dropping a restricted main promotes a public one: the topic would become
	// world-readable, so an author must not be able to do it.
	if _, err := ValidateTopicCategoryWrite(creator, everyone, TopicCategoryWrite{Current: []uint64{3, 1}, Next: []uint64{1}}); !errors.Is(err, ErrCategoryPermissionDenied) {
		t.Fatalf("author widening a restricted topic = %v, want denied", err)
	}
	if got, err := ValidateTopicCategoryWrite(manager, everyone, TopicCategoryWrite{Current: []uint64{3, 1}, Next: []uint64{1}}); err != nil || MainCategoryOf(got) != 1 {
		t.Fatalf("manager widening = %v, %v, want main 1", got, err)
	}
	// The last category can never be removed: there would be no main category
	// left and the topic would be readable by nobody.
	if _, err := ValidateTopicCategoryWrite(creator, everyone, TopicCategoryWrite{Current: []uint64{1}, Next: []uint64{}}); !errors.Is(err, ErrCategoryRequired) {
		t.Fatalf("removing the last category = %v, want ErrCategoryRequired", err)
	}
}

func TestGlobalManagerBypassesCategoryMap(t *testing.T) {
	store := newFakeStore()
	resolver := NewResolver(store, func(userID uint64) bool { return userID == 9 }, nil)
	snapshot, err := resolver.Resolve(9)
	if err != nil {
		t.Fatalf("Resolve manager: %v", err)
	}
	if !snapshot.HasGlobalManage() || !snapshot.CanManageCategory(999999) {
		t.Fatal("global manager did not bypass category grants")
	}
}

func TestResolverCachesAndPreciselyInvalidatesMetadata(t *testing.T) {
	store := newFakeStore()
	store.memberships[7] = []uint64{20}
	store.grants[10] = []CategoryGrant{{CategoryID: 1, Capability: CapabilityRead}}
	store.grants[11] = []CategoryGrant{{CategoryID: 2, Capability: CapabilityCreate}}
	store.grants[20] = []CategoryGrant{{CategoryID: 3, Capability: CapabilityRead}}
	resolver := NewResolver(store, nil, nil)

	if _, err := resolver.Resolve(7); err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	if _, err := resolver.Resolve(7); err != nil {
		t.Fatalf("second Resolve: %v", err)
	}
	if store.memberLoads != 1 || store.grantLoads[20] != 1 {
		t.Fatalf("cache loads members=%d grants=%v", store.memberLoads, store.grantLoads)
	}

	resolver.InvalidateUser(7)
	resolver.InvalidateGroup(20)
	if _, err := resolver.Resolve(7); err != nil {
		t.Fatalf("Resolve after invalidation: %v", err)
	}
	if store.memberLoads != 2 || store.grantLoads[20] != 2 || store.grantLoads[10] != 1 || store.grantLoads[11] != 1 {
		t.Fatalf("precise invalidation loads members=%d grants=%v", store.memberLoads, store.grantLoads)
	}
}

func TestResolverFailsClosedWhenRequiredSystemGroupsAreMissing(t *testing.T) {
	store := newFakeStore()
	delete(store.system, accessGroups.SystemKeyRegistered)
	resolver := NewResolver(store, nil, nil)
	if _, err := resolver.Resolve(0); err == nil {
		t.Fatal("Resolve missing system group error = nil")
	}
}

func TestResolverRejectsMembershipOverflow(t *testing.T) {
	store := newFakeStore()
	for id := uint64(100); id < 100+MaxActiveCustomGroups+1; id++ {
		store.memberships[7] = append(store.memberships[7], id)
	}
	resolver := NewResolver(store, nil, nil)
	_, err := resolver.Resolve(7)
	if !errors.Is(err, ErrTooManyActiveGroups) {
		t.Fatalf("Resolve overflow error = %v", err)
	}
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		system: map[string]uint64{
			accessGroups.SystemKeyEveryone:   10,
			accessGroups.SystemKeyRegistered: 11,
		},
		memberships: make(map[uint64][]uint64),
		grants:      make(map[uint64][]CategoryGrant),
		grantLoads:  make(map[uint64]int),
	}
}

func assertCapability(t *testing.T, snapshot Snapshot, categoryID uint64, want Capability) {
	t.Helper()
	if got := snapshot.Capability(categoryID); got != want {
		t.Fatalf("Capability(%d) = %d, want %d", categoryID, got, want)
	}
}
