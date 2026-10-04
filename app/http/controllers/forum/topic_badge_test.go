package forum

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
)

func TestBuildParticipantsPreservesWornBadges(t *testing.T) {
	moderator := &badgeservice.UserBadge{Code: "moderator"}
	robot := &badgeservice.UserBadge{Code: "robot"}
	topic := &vo.TopicsSimpleVo{
		AuthorId:  1,
		Username:  "author",
		WornBadge: moderator,
		Posters:   []vo.PosterVo{{Id: 2, Username: "poster", WornBadge: robot}},
	}

	participants := buildParticipants(topic)
	if len(participants) != 2 {
		t.Fatalf("participants length = %d, want 2", len(participants))
	}
	if participants[0].WornBadge == nil || participants[0].WornBadge.Code != "robot" {
		t.Fatalf("poster worn badge = %#v", participants[0].WornBadge)
	}
	if participants[1].WornBadge == nil || participants[1].WornBadge.Code != "moderator" {
		t.Fatalf("author worn badge = %#v", participants[1].WornBadge)
	}
}
