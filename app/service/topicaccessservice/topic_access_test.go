package topicaccessservice

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

func TestAuthorizePublishedRejectsUnavailableTopicBeforePermissionLookup(t *testing.T) {
	if err := AuthorizePublished(0, topics.Entity{Id: 7, Status: 0, ProcessStatus: 0}, accesscontrol.CapabilityRead); err != ErrUnavailable {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestAuthorizePublishedRejectsHiddenPublishedTopic(t *testing.T) {
	if err := AuthorizePublished(0, topics.Entity{Id: 7, Status: 1, ProcessStatus: 1}, accesscontrol.CapabilityRead); err != ErrUnavailable {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
