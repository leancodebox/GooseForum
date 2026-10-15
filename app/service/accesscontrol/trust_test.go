package accesscontrol

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

type trustFake struct {
	*batchFakeStore
	identities []users.TrustIdentity
}

func (s *trustFake) TrustLevels(ids []uint64) ([]users.TrustIdentity, error) {
	var rows []users.TrustIdentity
	for _, id := range ids {
		for _, u := range s.identities {
			if u.Id == id {
				rows = append(rows, u)
			}
		}
	}
	return rows, nil
}
func TestTrustGroupsAndNotificationAudiences(t *testing.T) {
	base := newFakeStore()
	base.system["trust:100"] = 100
	base.system["trust:200"] = 200
	base.system["trust:1000"] = 1000
	base.grants[100] = []CategoryGrant{{CategoryID: 9, Capability: CapabilityRead}}
	base.grants[200] = []CategoryGrant{{CategoryID: 10, Capability: CapabilityCreate}}
	s := &trustFake{batchFakeStore: &batchFakeStore{fakeStore: base}, identities: []users.TrustIdentity{{Id: 1, TrustLevel: trustlevel.Newcomer}, {Id: 2, TrustLevel: trustlevel.Member}, {Id: 3, TrustLevel: trustlevel.Core}, {Id: 4, TrustLevel: 30000}}}
	r := NewResolver(s, nil, nil)
	for _, tc := range []struct {
		id           uint64
		read, create bool
	}{{1, false, false}, {2, true, false}, {3, true, true}} {
		snap, err := r.Resolve(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		if snap.CanReadCategory(9) != tc.read || snap.CanCreateCategory(10) != tc.create {
			t.Fatalf("wrong grants for %d", tc.id)
		}
		audience, cache := snap.ListCacheAudience()
		if !cache || audience == "registered" {
			t.Fatal("trust audience not isolated")
		}
	}
	got, err := r.FilterReadableUserIDs([]uint64{4, 3, 2, 1}, 10)
	if err != nil || len(got) != 1 || got[0] != 3 {
		t.Fatalf("notification leak: %v %v", got, err)
	}
	if _, err := r.Resolve(4); err == nil {
		t.Fatal("unknown level accepted")
	}
}
