package userFollow

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestOrderUsersByIDs(t *testing.T) {
	input := []*users.EntityComplete{
		{Id: 3, Username: "three"},
		{Id: 1, Username: "one"},
		{Id: 2, Username: "two"},
	}

	got := orderUsersByIDs(input, []uint64{2, 3, 1, 4})
	if len(got) != 3 {
		t.Fatalf("len(orderUsersByIDs) = %d, want 3", len(got))
	}
	if got[0].Id != 2 || got[1].Id != 3 || got[2].Id != 1 {
		t.Fatalf("orderUsersByIDs ids = [%d %d %d], want [2 3 1]", got[0].Id, got[1].Id, got[2].Id)
	}
}
