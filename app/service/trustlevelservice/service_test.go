package trustlevelservice

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestPromotionAndBoundedProgress(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	c := pageConfig.DefaultTrustSettings()
	s := users.TrustState{AccountState: users.AccountState{CreatedAt: now.Add(-30 * time.Hour)}}
	for id := uint64(1); id <= 5; id++ {
		s = Next(s, c, now, id)
	}
	if s.TrustLevel != trustlevel.Newcomer || s.TrustVisitDays != 1 {
		t.Fatalf("premature promotion: %+v", s)
	}
	s = Next(s, c, now.Add(24*time.Hour), 1)
	if s.TrustLevel != trustlevel.Member || s.TrustVisitDays != 2 || len(s.TrustTopicIds) != 5 {
		t.Fatalf("member progress: %+v", s)
	}
	old := s
	s = Next(s, c, now, 1)
	if s.TrustVisitDays != old.TrustVisitDays || s.TrustLastVisitDay != old.TrustLastVisitDay {
		t.Fatal("out of order visit counted")
	}
	c.PromotionRules[1].Enabled = false
	for i := 0; i < 35; i++ {
		s = Next(s, c, now.Add(time.Duration(i+2)*24*time.Hour), uint64(i+6))
	}
	if s.TrustVisitDays <= 7 || len(s.TrustTopicIds) != 20 || s.TrustLevel != trustlevel.Member {
		t.Fatalf("bounded collection: %+v", s)
	}
	c.PromotionRules[1].Enabled = true
	s = Next(s, c, now.Add(40*24*time.Hour), 0)
	if s.TrustLevel != trustlevel.Regular {
		t.Fatal("time-only review did not promote")
	}
	c.PromotionRules[1].MinVisitDays = 36500
	if Next(s, c, now.Add(50*24*time.Hour), 99).TrustLevel != trustlevel.Regular {
		t.Fatal("conditions lowered existing level")
	}
	s.TrustLevel = trustlevel.Newcomer
	s.TrustLevelLocked = true
	if Next(s, c, now.Add(50*24*time.Hour), 99).TrustVisitDays != s.TrustVisitDays {
		t.Fatal("locked progress changed")
	}
}
func TestContentCountsMatchRenderedContent(t *testing.T) {
	c := CountContent("![a](/a.png) ![b](/a.png) ![remote](https://example.com/a.png)\n\n`![code](/b.png)`\n\n```\n![code](/c.png)\n```\n\n[link](https://example.com) [local](/local)", 0)
	if c.Images != 3 || c.Links != 1 {
		t.Fatalf("counts=%+v", c)
	}
}
