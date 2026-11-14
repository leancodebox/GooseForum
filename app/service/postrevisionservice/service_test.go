package postrevisionservice

import (
	"errors"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

type revisionAccessStore struct {
	grants map[uint64][]accesscontrol.CategoryGrant
}

func (store revisionAccessStore) SystemGroupIDs() (map[string]uint64, error) {
	return map[string]uint64{accessGroups.SystemKeyEveryone: 1, accessGroups.SystemKeyRegistered: 2}, nil
}
func (store revisionAccessStore) ActiveCustomGroupIDs(uint64) ([]uint64, error) { return nil, nil }
func (store revisionAccessStore) EnabledCategoryGrants(groupID uint64) ([]accesscontrol.CategoryGrant, error) {
	return store.grants[groupID], nil
}

func TestAuthorizeAllowsPostAuthorWithoutTopicVisibilityShortcut(t *testing.T) {
	post := posts.Entity{Id: 7, UserId: 42, ProcessStatus: 1}
	topic := topics.Entity{Id: 8, MainCategoryId: 9, Status: 0}
	if err := Authorize(42, post, topic); err != nil {
		t.Fatalf("author authorization = %v", err)
	}
}

func TestAuthorizeRejectsMissingResource(t *testing.T) {
	if err := Authorize(42, posts.Entity{}, topics.Entity{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing resource error = %v, want ErrNotFound", err)
	}
}

func TestAuthorizeRequiresManageScopeForNonAuthor(t *testing.T) {
	store := revisionAccessStore{grants: map[uint64][]accesscontrol.CategoryGrant{
		1: {{CategoryID: 9, Capability: accesscontrol.CapabilityRead}},
		2: {{CategoryID: 9, Capability: accesscontrol.CapabilityRead}},
	}}
	resolver := accesscontrol.NewResolver(store, func(uint64) bool { return false }, func(uint64) (bool, []uint64) { return false, nil })
	post := posts.Entity{Id: 7, UserId: 42}
	topic := topics.Entity{Id: 8, MainCategoryId: 9}
	if err := AuthorizeWithResolver(resolver, 43, post, topic); !errors.Is(err, ErrForbidden) {
		t.Fatalf("ordinary reader error = %v, want ErrForbidden", err)
	}
	store.grants[2] = []accesscontrol.CategoryGrant{{CategoryID: 9, Capability: accesscontrol.CapabilityManage}}
	resolver = accesscontrol.NewResolver(store, func(uint64) bool { return false }, func(uint64) (bool, []uint64) { return false, nil })
	if err := AuthorizeWithResolver(resolver, 43, post, topic); err != nil {
		t.Fatalf("category manager authorization = %v", err)
	}
}

func TestAuthorizeDoesNotUseManageScopeFromAnotherCategory(t *testing.T) {
	store := revisionAccessStore{grants: map[uint64][]accesscontrol.CategoryGrant{
		1: {{CategoryID: 11, Capability: accesscontrol.CapabilityManage}},
		2: {{CategoryID: 11, Capability: accesscontrol.CapabilityManage}},
	}}
	resolver := accesscontrol.NewResolver(store, func(uint64) bool { return false }, func(uint64) (bool, []uint64) { return false, nil })
	if err := AuthorizeWithResolver(resolver, 43, posts.Entity{Id: 7, UserId: 42}, topics.Entity{Id: 8, MainCategoryId: 10}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("cross-category manager error = %v, want ErrForbidden", err)
	}
}
