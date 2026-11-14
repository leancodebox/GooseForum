package postrevisionservice

import (
	"errors"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

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

func TestAuthorizeRejectsEveryNonAuthor(t *testing.T) {
	post := posts.Entity{Id: 7, UserId: 42}
	topic := topics.Entity{Id: 8, MainCategoryId: 9}
	for _, userID := range []uint64{0, 1, 43} {
		if err := Authorize(userID, post, topic); !errors.Is(err, ErrForbidden) {
			t.Fatalf("user %d authorization = %v, want ErrForbidden", userID, err)
		}
	}
	post.UserId = 0
	if err := Authorize(0, post, topic); !errors.Is(err, ErrForbidden) {
		t.Fatalf("anonymous author authorization = %v, want ErrForbidden", err)
	}
}
