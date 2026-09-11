package api

import (
	"fmt"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/contentmoderationservice"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
	"github.com/leancodebox/GooseForum/app/service/topicservice"
)

func setReviewMode(t *testing.T, mode string) {
	t.Helper()
	if err := pageConfig.SaveConfig(pageConfig.SensitiveWordSettings, jsonopt.Encode(pageConfig.SensitiveWordConfig{Enabled: true, Mode: mode})); err != nil {
		t.Fatal(err)
	}
	sensitivewordservice.ClearConfigCache()
}

func TestAsyncTopicVisibilityAndReplacement(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	enableReviewTest(t)
	replacement := sensitiveWord.Entity{Word: "replace-review-test", Action: "replace", Replacement: "replacement-visible", Enabled: true}
	if err := sensitiveWord.Save(&replacement); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sensitiveWord.Delete(replacement.Id); sensitivewordservice.Refresh() })
	sensitivewordservice.Refresh()
	cases := []struct {
		mode, content string
		before, after int8
	}{
		{"after_review", "ordinary content", 0, 1},
		{"after_review", "blocked-review-test", 0, 0},
		{"after_review", "replace-review-test", 0, 1},
		{"visible_then_review", "ordinary content", 1, 1},
		{"visible_then_review", "blocked-review-test", 1, 0},
		{"visible_then_review", "replace-review-test", 1, 1},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%s/%d", tc.mode, i), func(t *testing.T) {
			setReviewMode(t, tc.mode)
			id := uint64(950000 + i*10)
			seedAdminTopic(t, db, id)
			topic := topics.Get(id)
			post := posts.Get(topic.FirstPostId)
			post.Content = tc.content
			contentmoderationservice.PrepareTopic(&topic, &post)
			if topic.Status != tc.before || topic.ModerationStatus != "pending" || post.Content != tc.content {
				t.Fatal("prepare must only set pending visibility, not run word matching")
			}
			if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds}); err != nil {
				t.Fatal(err)
			}
			if err := runContentReview(contentReviewJob{topic: true, id: id, version: topic.ModerationVersion}); err != nil {
				t.Fatal(err)
			}
			stored := topics.Get(id)
			body := posts.Get(post.Id)
			if stored.Status != tc.after || stored.ModerationStatus == "pending" {
				t.Fatalf("review result: %+v", stored)
			}
			if tc.content == "replace-review-test" && (body.Content != "replacement-visible" || stored.Excerpt != "replacement-visible") {
				t.Fatal("replacement or topic excerpt not persisted")
			}
		})
	}
}

func TestOldJobCannotPublishDraftOrOverwriteManualDecision(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	enableReviewTest(t)
	seedAdminTopic(t, db, 951000)
	topic := topics.Get(951000)
	post := posts.Get(topic.FirstPostId)
	contentmoderationservice.PrepareTopic(&topic, &post)
	job := contentReviewJob{topic: true, id: topic.Id, version: topic.ModerationVersion}
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds}); err != nil {
		t.Fatal(err)
	}
	topic.Status = 0
	contentmoderationservice.PrepareTopic(&topic, &post)
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds}); err != nil {
		t.Fatal(err)
	}
	if err := runContentReview(job); err != nil {
		t.Fatal(err)
	}
	if got := topics.Get(topic.Id); got.Status != 0 || got.ModerationStatus != "none" {
		t.Fatal("stale worker published a draft")
	}
}

func TestAsyncReplyVisibilityAndPublicCounters(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	enableReviewTest(t)
	for i, mode := range []string{"after_review", "visible_then_review"} {
		t.Run(mode, func(t *testing.T) {
			setReviewMode(t, mode)
			id := uint64(952000 + i*10)
			seedAdminTopic(t, db, id)
			topic := topics.Get(id)
			post := posts.Entity{TopicId: id, UserId: topic.UserId, Content: "blocked-review-test"}
			if err := postservice.CreateTopicPost(&post, topic); err != nil {
				t.Fatal(err)
			}
			before := int8(1)
			if mode == "visible_then_review" {
				before = 0
			}
			if post.ProcessStatus != before || post.ModerationStatus != "pending" {
				t.Fatal("incorrect reply visibility before worker")
			}
			if err := runContentReview(contentReviewJob{id: post.Id, version: post.ModerationVersion}); err != nil {
				t.Fatal(err)
			}
			stored := posts.Get(post.Id)
			if stored.ProcessStatus != 1 || stored.ModerationStatus != "rejected" {
				t.Fatal("reply not rejected")
			}
			if got := topics.Get(id); got.ReplyCount != topic.ReplyCount {
				t.Fatalf("public reply count = %d, want %d", got.ReplyCount, topic.ReplyCount)
			}
		})
	}
}

func TestReviewWritePreservesConcurrentCountersAndManualBlock(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	enableReviewTest(t)
	seedAdminTopic(t, db, 953000)
	topic := topics.Get(953000)
	post := posts.Get(topic.FirstPostId)
	contentmoderationservice.PrepareTopic(&topic, &post)
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds}); err != nil {
		t.Fatal(err)
	}
	version := topic.ModerationVersion
	topic.Status = 1
	contentmoderationservice.ReviewTopic(&topic, &post)
	db.Model(&topics.Entity{}).Where("id = ?", topic.Id).Updates(map[string]any{"reply_count": 17, "pin_weight": 8, "process_status": 1})
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds, ExpectedVersion: &version, ReviewOnly: true}); err != nil {
		t.Fatal(err)
	}
	got := topics.Get(topic.Id)
	if got.ReplyCount != 17 || got.PinWeight != 8 || got.ProcessStatus != 1 {
		t.Fatal("worker overwrote unrelated changes")
	}
}

func TestManualReplyBlockInvalidatesPendingJob(t *testing.T) {
	db := setupTopicWriteTestDB(t)
	enableReviewTest(t)
	seedAdminTopic(t, db, 954000)
	topic := topics.Get(954000)
	post := posts.Entity{TopicId: topic.Id, UserId: topic.UserId, Content: "ordinary content"}
	if err := postservice.CreateTopicPost(&post, topic); err != nil {
		t.Fatal(err)
	}
	job := contentReviewJob{id: post.Id, version: post.ModerationVersion}
	if err := posts.UpdateProcessStatus(post.Id, 1); err != nil {
		t.Fatal(err)
	}
	if err := runContentReview(job); err != nil {
		t.Fatal(err)
	}
	stored := posts.Get(post.Id)
	if stored.ProcessStatus != 1 || stored.ModerationStatus == "pending" {
		t.Fatal("worker undid manual block")
	}
	contentmoderationservice.PreparePost(&stored)
	if stored.ProcessStatus != 1 || stored.ModerationStatus == "pending" {
		t.Fatal("editing a manually blocked reply queued automatic release")
	}
}
