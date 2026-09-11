package contentmoderationservice

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
	"strings"
	"testing"
)

func TestReviewPreservesDraftAndChecksEdits(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}, &sensitiveWord.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := pageConfig.SaveConfig(pageConfig.SensitiveWordSettings, `{"enabled":true}`); err != nil {
		t.Fatal(err)
	}
	sensitivewordservice.ClearConfigCache()
	for _, word := range []sensitiveWord.Entity{{Word: "forbidden", Action: "reject", Enabled: true}, {Word: "replace-me", Action: "replace", Replacement: "safe", Enabled: true}} {
		w := word
		if err := sensitiveWord.Save(&w); err != nil {
			t.Fatal(err)
		}
	}
	sensitivewordservice.Refresh()
	topic := topics.Entity{Title: "draft", Status: 0}
	post := posts.Entity{Content: "forbidden"}
	ReviewTopic(&topic, &post)
	if topic.Status != 0 || topic.ModerationStatus != "none" {
		t.Fatalf("draft changed: %+v", topic)
	}
	topic.Status = 1
	ReviewTopic(&topic, &post)
	if topic.Status != 0 || topic.ModerationStatus != "rejected" {
		t.Fatal("publication bypassed rejection")
	}
	topic.Status = 1
	post.Content = "replace-me"
	ReviewTopic(&topic, &post)
	if topic.Status != 1 || post.Content != "safe" || !strings.Contains(post.RenderedHTML, "safe") || topic.Excerpt != "safe" {
		t.Fatal("replacement projections not updated")
	}
	version := topic.ModerationVersion
	topic.Title = "forbidden"
	ReviewTopic(&topic, &post)
	if topic.Status != 0 || topic.ModerationVersion != version+1 {
		t.Fatal("edit bypassed review")
	}
	reply := posts.Entity{Content: "forbidden"}
	ReviewPost(&reply)
	if reply.ProcessStatus != 1 || reply.PublishedAt != nil {
		t.Fatal("reply not hidden")
	}
	reply.Content = "safe"
	ReviewPost(&reply)
	if reply.ProcessStatus != 0 {
		t.Fatal("corrected reply remains hidden")
	}
	publishedAt := reply.PublishedAt
	if publishedAt == nil {
		t.Fatal("first publication not recorded")
	}
	reply.Content = "forbidden"
	ReviewPost(&reply)
	reply.Content = "safe"
	ReviewPost(&reply)
	if reply.PublishedAt != publishedAt {
		t.Fatal("restoration reset first publication")
	}
	reply.ProcessStatus = 1
	ReviewPost(&reply)
	if reply.ProcessStatus != 1 {
		t.Fatal("manual block cleared")
	}
}
