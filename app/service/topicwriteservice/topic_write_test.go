package topicwriteservice

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/markdownext"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/mentionservice"
)

type legacyTopicAccessStore struct{ categoryID uint64 }

func (legacyTopicAccessStore) SystemGroupIDs() (map[string]uint64, error) {
	return map[string]uint64{"everyone": 1, "registered": 2}, nil
}

func (legacyTopicAccessStore) ActiveCustomGroupIDs(uint64) ([]uint64, error) {
	return nil, nil
}

func (store legacyTopicAccessStore) EnabledCategoryGrants(uint64) ([]accesscontrol.CategoryGrant, error) {
	return []accesscontrol.CategoryGrant{{CategoryID: store.categoryID, Capability: accesscontrol.CapabilityCreate}}, nil
}

func TestPrepareWritePreservesMentionsFromPreviouslyPublishedLegacyTopics(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&category.Entity{}, &users.EntityComplete{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	group := category.Entity{Name: "Legacy mention regression"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&group) })
	previous := accesscontrol.Default
	accesscontrol.Default = accesscontrol.NewResolver(legacyTopicAccessStore{group.Id}, nil, nil)
	t.Cleanup(func() { accesscontrol.Default = previous })
	publishedAt := time.Now().Add(-time.Hour)
	oldMention := markdownext.Canonical(98731001, "old-recipient")
	newMention := markdownext.Canonical(98731002, "new-recipient")
	for _, tc := range []struct {
		name        string
		status      int8
		hidden      int8
		publishedAt *time.Time
	}{
		{"hidden-published-topic", 1, 1, &publishedAt},
		{"withdrawn-published-topic", 0, 0, &publishedAt},
		{"never-published-draft", 0, 0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &writeState{
				topic:     topics.Entity{Id: 98731101, Status: tc.status, ProcessStatus: tc.hidden, PublishedAt: tc.publishedAt, CategoryIds: []uint64{group.Id}},
				firstPost: posts.Entity{Id: 98731201, Content: oldMention, SourceVersion: 0},
			}
			_, err := prepareWrite(state, WriteInput{UserID: 98731301, Status: 1, SourceVersion: 1, Content: oldMention + " " + newMention, CategoryIDs: []uint64{group.Id}})
			if err != nil {
				t.Fatal(err)
			}
			recipients := mentionservice.Recipients(state.firstPost)
			if tc.publishedAt != nil {
				if len(recipients) != 1 || recipients[0] != 98731002 {
					t.Fatalf("historical recipient notified: %v", recipients)
				}
			} else if len(recipients) != 2 || len(state.firstPost.LegacyMentionIDs) != 0 {
				t.Fatalf("first publication suppressed recipients: %v", recipients)
			}
		})
	}
}
