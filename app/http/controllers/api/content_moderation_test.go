package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
)

type contentReviewJob struct {
	topic       bool
	id, version uint64
}

func runContentReview(job contentReviewJob) error {
	event := &eventhandlers.ContentReviewRequestedEvent{Topic: job.topic, ID: job.id, Version: job.version}
	for _, handler := range eventhandlers.Handlers() {
		if _, ok := handler.NewEvent().(*eventhandlers.ContentReviewRequestedEvent); ok {
			return handler.Handle(context.Background(), event)
		}
	}
	return fmt.Errorf("content review event handler is not registered")
}

func runPendingTopicReview(t *testing.T, id uint64) {
	t.Helper()
	topic := topics.Get(id)
	if err := runContentReview(contentReviewJob{topic: true, id: id, version: topic.ModerationVersion}); err != nil {
		t.Fatal(err)
	}
}

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

func TestAdminReviewBatchKeepsPerItemVersionConflicts(t *testing.T) {
	db := setupAdminTopicTestDB(t)
	seedAdminTopic(t, db, 940011)
	seedAdminTopic(t, db, 940012)
	result := ReviewAdminTopicBatch(component.BetterRequest[ReviewBatchReq]{UserId: 1, Params: ReviewBatchReq{
		Items:  []ReviewBatchItem{{Id: 940011, Version: 0}, {Id: 940012, Version: 99}},
		Action: "reject",
		Reason: "batch review",
	}})
	batch, ok := result.Data.Result.(ReviewBatchResult)
	if !ok || len(batch.Items) != 2 || !batch.Items[0].Success || batch.Items[1].Success {
		t.Fatalf("unexpected batch result: %#v", result.Data.Result)
	}
	if got := topics.Get(940011); got.Status != 0 || got.ModerationStatus != "rejected" {
		t.Fatalf("successful item was not rejected: %+v", got)
	}
	if got := topics.Get(940012); got.Status != 1 || got.ModerationStatus == "rejected" {
		t.Fatalf("stale item changed unexpectedly: %+v", got)
	}
}

func TestAdminReviewBatchRejectsDuplicateItems(t *testing.T) {
	result := ReviewAdminTopicBatch(component.BetterRequest[ReviewBatchReq]{Params: ReviewBatchReq{
		Items:  []ReviewBatchItem{{Id: 1, Version: 0}, {Id: 1, Version: 0}},
		Action: "approve",
		Reason: "duplicate",
	}})
	if result.Data.Code != component.FAIL {
		t.Fatalf("duplicate batch should fail: %+v", result)
	}
}

func TestAdminReplyReviewBatchKeepsPerItemVersionConflicts(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	userID := uint64(950101)
	topicID := uint64(950001)
	firstPostID := uint64(950011)
	replyID := uint64(950012)
	staleReplyID := uint64(950013)
	createTopicWriteUser(t, db, userID, "batch-reply")
	now := time.Now().Add(-time.Hour)
	topic := topics.Entity{Id: topicID, Title: "Batch reply review", UserId: userID, Status: 1, PostCount: 3, ReplyCount: 2, PostSeq: 3, FirstPostId: firstPostID, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&topic).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}
	for _, post := range []posts.Entity{
		{Id: firstPostID, TopicId: topicID, PostNo: 1, UserId: userID, Content: "first post", CreatedAt: now, UpdatedAt: now},
		{Id: replyID, TopicId: topicID, PostNo: 2, UserId: userID, Content: "reply one", ProcessStatus: 1, ModerationStatus: "pending", CreatedAt: now, UpdatedAt: now},
		{Id: staleReplyID, TopicId: topicID, PostNo: 3, UserId: userID, Content: "reply two", ProcessStatus: 1, ModerationStatus: "pending", ModerationVersion: 2, CreatedAt: now, UpdatedAt: now},
	} {
		if err := db.Create(&post).Error; err != nil {
			t.Fatalf("create post %d: %v", post.Id, err)
		}
	}
	result := ReviewAdminPostBatch(component.BetterRequest[ReviewBatchReq]{UserId: 1, Params: ReviewBatchReq{
		Items:  []ReviewBatchItem{{Id: replyID, Version: 0}, {Id: staleReplyID, Version: 0}},
		Action: "approve",
		Reason: "batch reply review",
	}})
	batch, ok := result.Data.Result.(ReviewBatchResult)
	if !ok || len(batch.Items) != 2 || !batch.Items[0].Success || batch.Items[1].Success {
		t.Fatalf("unexpected reply batch result: %#v", result.Data.Result)
	}
	if got := posts.Get(replyID); got.ProcessStatus != 0 || got.ModerationStatus != "approved" {
		t.Fatalf("successful reply was not approved: %+v", got)
	}
	if got := posts.Get(staleReplyID); got.ProcessStatus != 1 || got.ModerationStatus != "pending" || got.ModerationVersion != 2 {
		t.Fatalf("stale reply changed unexpectedly: %+v", got)
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
	if err := runContentReview(contentReviewJob{id: post.Id, version: post.ModerationVersion}); err != nil {
		t.Fatal(err)
	}
	rows := ReviewPostsList(component.BetterRequest[ReviewPostsReq]{Params: ReviewPostsReq{ModerationStatus: "rejected", PageSize: 50}}).Data.Result.(component.Page[ReviewPostVo])
	if len(rows.List) == 0 {
		t.Fatal("rejected reply missing from queue")
	}
}

func TestPublishingAndEditingEnqueueReview(t *testing.T) {
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
	runPendingTopicReview(t, id)
	if topic := topics.Get(id); topic.Status != 0 || topic.ModerationStatus != "rejected" {
		t.Fatal("draft publication bypassed review")
	}
	req.Params.TopicId = id
	req.Params.TopicStatus = 1
	req.Params.Content = "Normal content with enough words"
	WriteTopic(req)
	runPendingTopicReview(t, id)
	if topics.Get(id).Status != 1 {
		t.Fatal("corrected topic not published")
	}
	req.Params.Content = "blocked-review-test content"
	WriteTopic(req)
	runPendingTopicReview(t, id)
	if topic := topics.Get(id); topic.Status != 0 || topic.ModerationStatus != "rejected" {
		t.Fatal("editing bypassed review")
	}
}
