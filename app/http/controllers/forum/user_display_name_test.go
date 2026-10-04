package forum

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestTopicUserNamesPreserveAccountIdentity(t *testing.T) {
	user := &users.EntityComplete{Id: 7, Username: "account", Nickname: " Display name "}
	author := userPayloadWithWornBadge(7, map[uint64]*users.EntityComplete{7: user}, nil)
	if author.Username != "account" || author.Nickname != user.Nickname || author.DisplayName() != "Display name" {
		t.Fatalf("author identity = %#v", author)
	}
	participants := buildParticipants(&vo.TopicsSimpleVo{
		AuthorId: 7, Username: "account", Nickname: "Display name",
		Posters: []vo.PosterVo{{Id: 8, Username: "reply-account", Nickname: "Reply name"}},
	})
	if len(participants) != 2 || participants[0].Nickname != "Reply name" || participants[1].Nickname != "Display name" {
		t.Fatalf("participants = %#v", participants)
	}
	for _, nickname := range []string{"", " \t "} {
		if got := (TopicAuthorPayload{Username: "account", Nickname: nickname}).DisplayName(); got != "account" {
			t.Fatalf("blank nickname should fall back to account, got %q", got)
		}
	}
}
