package mentionservice

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/markdownext"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

func createMentionUsers(t *testing.T) []users.EntityComplete {
	t.Helper()
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&users.EntityComplete{}, &eventNotification.Entity{}); err != nil {
		t.Fatal(err)
	}
	identities := []users.EntityComplete{{Id: 970001, Username: "mention_alice"}, {Id: 970002, Username: "mention_bob"}, {Id: 970003, Username: "mention_frozen", RestrictionStatus: users.RestrictionSuspended}, {Id: 970004, Username: "mention_extra"}, {Id: 970005, Username: "mention_other"}}
	if err := conn.Create(&identities).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn.Unscoped().Where("id IN ?", []uint64{970001, 970002, 970003, 970004, 970005}).Delete(&users.EntityComplete{})
		conn.Where("topic_id = ?", 970101).Delete(&eventNotification.Entity{})
	})
	return identities
}

func TestNormalizeStableIdentityAndProtectedSource(t *testing.T) {
	users := createMentionUsers(t)
	source := "你好 @mention_alice. **@mention_bob** `@mention_alice` [@mention_bob](/url)\\@mention_alice\n\n[future]\n@mention_alice\n[/future]"
	content, err := Normalize(source, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, markdownext.Canonical(users[0].Id, users[0].Username)) || !strings.Contains(content, markdownext.Canonical(users[1].Id, users[1].Username)) {
		t.Fatalf("not canonical: %s", content)
	}
	for _, want := range []string{"`@mention_alice`", "[@mention_bob](/url)", "\\@mention_alice", "[future]\n@mention_alice\n[/future]"} {
		if !strings.Contains(content, want) {
			t.Fatalf("changed protected %q: %s", want, content)
		}
	}
	legacy, err := Normalize(source, 0, true)
	if err != nil || legacy != source {
		t.Fatal("legacy source changed")
	}
	unknown, err := Normalize("@missing_user @missing_other", 1, true)
	if err != nil || unknown != "@missing_user @missing_other" {
		t.Fatalf("unknown identity resolved: %q %v", unknown, err)
	}
	wrongCase, err := Normalize("@MENTION_ALICE", 1, true)
	if err != nil || wrongCase != markdownext.Canonical(users[0].Id, users[0].Username) {
		t.Fatalf("case-folded identity not resolved: %q %v", wrongCase, err)
	}
	if _, err := Normalize("[mention user=\"bad\"]@alice[/mention]", 1, true); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid accepted %v", err)
	}
}

func TestHydrateRenameDeleteAndBatch(t *testing.T) {
	identities := createMentionUsers(t)
	conn := dbconnect.Connect()
	if err := conn.Model(&users.EntityComplete{}).Where("id = ?", identities[0].Id).Update("username", "mention_renamed").Error; err != nil {
		t.Fatal(err)
	}
	raw := fmt.Sprintf(`<p><a class="mention" data-mention-user="%d" href="/u/%d">@old</a></p>`, identities[0].Id, identities[0].Id)
	values := HydrateHTMLs([]string{raw, "<p>plain</p>", raw})
	if !strings.Contains(values[0], "@mention_renamed") || !strings.Contains(values[2], "@mention_renamed") || values[1] != "<p>plain</p>" {
		t.Fatalf("hydration failed %#v", values)
	}
	if err := conn.Delete(&users.EntityComplete{}, identities[0].Id).Error; err != nil {
		t.Fatal(err)
	}
	if got := HydrateHTML(raw); strings.Contains(got, "href=") || !strings.Contains(got, "@old") {
		t.Fatalf("deleted identity not downgraded: %s", got)
	}
}

func TestLegacyUpgradeSuppressesOldRecipientsButAllowsNewMentions(t *testing.T) {
	identities := createMentionUsers(t)
	post := posts.Entity{Content: "Historical @mention_alice and `@mention_bob`", SourceVersion: 0}
	if err := PreserveLegacyMentions(&post, 1, true); err != nil {
		t.Fatal(err)
	}
	if len(post.LegacyMentionIDs) != 1 || post.LegacyMentionIDs[0] != identities[0].Id {
		t.Fatalf("baseline: %v", post.LegacyMentionIDs)
	}
	post.SourceVersion = 1
	post.Content = markdownext.Canonical(identities[0].Id, identities[0].Username) + " " + markdownext.Canonical(identities[1].Id, identities[1].Username)
	ids := Recipients(post)
	if len(ids) != 1 || ids[0] != identities[1].Id {
		t.Fatalf("recipients: %v", ids)
	}
	post.Content = ""
	post.Content = markdownext.Canonical(identities[0].Id, identities[0].Username)
	if len(Recipients(post)) != 0 {
		t.Fatal("later edits replayed a historical mention")
	}
	draft := posts.Entity{Content: "@mention_alice"}
	if err := PreserveLegacyMentions(&draft, 1, false); err != nil || len(draft.LegacyMentionIDs) != 0 {
		t.Fatal("unpublished draft suppressed mentions")
	}
}

type mentionAccessStore struct{}

func (mentionAccessStore) SystemGroupIDs() (map[string]uint64, error) {
	return map[string]uint64{"everyone": 1, "registered": 2}, nil
}
func (mentionAccessStore) ActiveCustomGroupIDs(id uint64) ([]uint64, error) {
	if id == 970002 {
		return []uint64{3}, nil
	}
	return nil, nil
}
func (mentionAccessStore) EnabledCategoryGrants(id uint64) ([]accesscontrol.CategoryGrant, error) {
	if id == 3 {
		return []accesscontrol.CategoryGrant{{CategoryID: 7, Capability: accesscontrol.CapabilityRead}}, nil
	}
	return nil, nil
}

func TestMentionVisibilityModerationAndNotificationDedupe(t *testing.T) {
	identities := createMentionUsers(t)
	previous := accesscontrol.Default
	accesscontrol.Default = accesscontrol.NewResolver(mentionAccessStore{}, nil, nil)
	t.Cleanup(func() { accesscontrol.Default = previous })
	topic := topics.Entity{Id: 970101, Status: 1, MainCategoryId: 7}
	post := posts.Entity{Id: 970201, TopicId: topic.Id, PostNo: 2, UserId: identities[0].Id, SourceVersion: 1, Content: markdownext.Canonical(identities[1].Id, identities[1].Username) + " " + markdownext.Canonical(identities[0].Id, identities[0].Username) + " " + markdownext.Canonical(identities[2].Id, identities[2].Username),
		ModerationStatus: "pending"}
	Notify(topic, post)
	var count int64
	conn := dbconnect.Connect()
	conn.Model(&eventNotification.Entity{}).Where("topic_id = ?", topic.Id).Count(&count)
	if count != 0 {
		t.Fatal("pending content sent notifications")
	}
	post.ModerationStatus = "approved"
	Notify(topic, post)
	Notify(topic, post)
	post.Content = "no mentions"
	Notify(topic, post)
	post.Content = markdownext.Canonical(identities[1].Id, identities[1].Username)
	Notify(topic, post)
	conn.Model(&eventNotification.Entity{}).Where("topic_id = ?", topic.Id).Count(&count)
	if count != 1 {
		t.Fatalf("count=%d want one eligible recipient once", count)
	}
	key := DedupeKey(post.Id+1, identities[1].Id)
	direct := eventNotification.Entity{TopicId: topic.Id, UserId: identities[1].Id, EventType: eventNotification.EventTypePostReply, DedupeKey: &key}
	if err := eventNotification.Create(&direct); err != nil {
		t.Fatal(err)
	}
	post.Id++
	Notify(topic, post)
	var notifications []eventNotification.Entity
	conn.Where("dedupe_key = ?", key).Find(&notifications)
	if len(notifications) != 1 || notifications[0].EventType != eventNotification.EventTypePostReply {
		t.Fatalf("direct reply not prioritized: %#v", notifications)
	}
}
