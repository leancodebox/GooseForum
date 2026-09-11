package api

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
	"testing"
)

func enableReviewTest(t *testing.T) {
	t.Helper()
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}, &sensitiveWord.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := pageConfig.SaveConfig(pageConfig.SensitiveWordSettings, `{"enabled":true}`); err != nil {
		t.Fatal(err)
	}
	word := sensitiveWord.Entity{Word: "blocked-review-test", Action: "reject", Enabled: true}
	if err := sensitiveWord.Save(&word); err != nil {
		t.Fatal(err)
	}
	sensitivewordservice.ClearConfigCache()
	sensitivewordservice.Refresh()
	t.Cleanup(func() {
		pageConfig.SaveConfig(pageConfig.SensitiveWordSettings, `{"enabled":false}`)
		sensitiveWord.Delete(word.Id)
		sensitivewordservice.ClearConfigCache()
		sensitivewordservice.Refresh()
	})
}

func TestAdminReviewRejectsStaleDecisionAndDoesNotPublishDraft(t *testing.T) {
	db := setupAdminTopicTestDB(t)
	seedAdminTopic(t, db, 940001)
	db.Model(&topics.Entity{}).Where("id = ?", 940001).Updates(map[string]any{"status": 0, "moderation_status": "rejected", "moderation_version": 2})
	ReviewAdminTopic(component.BetterRequest[ReviewContentReq]{Params: ReviewContentReq{Id: 940001, Version: 1, Action: "approve", Reason: "checked"}})
	if topics.Get(940001).Status != 0 {
		t.Fatal("stale decision published content")
	}
	ReviewAdminTopic(component.BetterRequest[ReviewContentReq]{Params: ReviewContentReq{Id: 940001, Version: 2, Action: "approve", Reason: "checked"}})
	got := topics.Get(940001)
	if got.Status != 1 || got.ModerationStatus != "approved" || got.ModerationVersion != 3 {
		t.Fatalf("approval failed: %+v", got)
	}
	db.Model(&topics.Entity{}).Where("id = ?", 940001).Updates(map[string]any{"status": 0, "moderation_status": "none"})
	ReviewAdminTopic(component.BetterRequest[ReviewContentReq]{Params: ReviewContentReq{Id: 940001, Version: 3, Action: "approve", Reason: "checked"}})
	if topics.Get(940001).Status != 0 {
		t.Fatal("draft published by administrator review")
	}
}

func TestReviewListFiltersStatusAndCategory(t *testing.T) {
	db := setupAdminTopicTestDB(t)
	_, categoryID := seedAdminTopic(t, db, 941001)
	db.Model(&topics.Entity{}).Where("id = ?", 941001).Update("moderation_status", "rejected")
	result := TopicsList(component.BetterRequest[TopicsListReq]{Params: TopicsListReq{ModerationStatus: "rejected", CategoryId: categoryID, PageSize: 10}})
	page := result.Data.Result.(component.Page[TopicInfoAdminVo])
	if len(page.List) != 1 || page.List[0].Id != 941001 || page.List[0].ModerationStatus != "rejected" {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestRejectedReplyDoesNotIncrementPublicStats(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	seedAdminTopic(t, db, 942001)
	enableReviewTest(t)
	topic := topics.Get(942001)
	post := posts.Entity{TopicId: topic.Id, UserId: topic.UserId, Content: "blocked-review-test"}
	if err := postservice.CreateTopicPost(&post, topic); err != nil {
		t.Fatal(err)
	}
	got := topics.Get(topic.Id)
	if post.ProcessStatus != 1 || got.ReplyCount != topic.ReplyCount || got.LastPostId != topic.LastPostId {
		t.Fatal("rejected reply updated public counters")
	}
	rows := ReviewPostsList(component.BetterRequest[ReviewPostsReq]{Params: ReviewPostsReq{ModerationStatus: "rejected", PageSize: 50}}).Data.Result.(component.Page[ReviewPostVo])
	if len(rows.List) == 0 {
		t.Fatal("rejected reply missing from queue")
	}
}

func TestPublishingAndEditingRunReviewBeforeSave(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	createTopicWriteUser(t, db, 943001, "review-writer")
	if err := db.Create(&category.Entity{Id: 943002, Name: "Review", Slug: "review-test"}).Error; err != nil {
		t.Fatal(err)
	}
	ensureTopicWriteAccess(t, db)
	enableReviewTest(t)
	req := component.BetterRequest[WriteTopicReq]{UserId: 943001, Params: WriteTopicReq{
		Title: "Review integration", Content: "blocked-review-test content", CategoryId: []uint64{943002}, TopicStatus: 0,
	}}
	result := WriteTopic(req)
	id, ok := result.Data.Result.(uint64)
	if !ok || id == 0 {
		t.Fatalf("draft save failed: %+v", result)
	}
	if topics.Get(id).Status != 0 {
		t.Fatal("draft published")
	}
	UpdateTopicStatus(component.BetterRequest[TopicStatusReq]{UserId: 943001, Params: TopicStatusReq{TopicId: id, TopicStatus: 1}})
	if topic := topics.Get(id); topic.Status != 0 || topic.ModerationStatus != "rejected" {
		t.Fatal("draft publication bypassed review")
	}
	req.Params.TopicId = id
	req.Params.TopicStatus = 1
	req.Params.Content = "Normal content with enough words"
	WriteTopic(req)
	if topics.Get(id).Status != 1 {
		t.Fatal("corrected topic not published")
	}
	req.Params.Content = "blocked-review-test content"
	WriteTopic(req)
	if topic := topics.Get(id); topic.Status != 0 || topic.ModerationStatus != "rejected" {
		t.Fatal("editing bypassed review")
	}
}
