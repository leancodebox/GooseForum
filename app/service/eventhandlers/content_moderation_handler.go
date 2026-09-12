package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/contentmoderationservice"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/topicservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

// ContentReviewRequestedEvent requests automatic review of one persisted
// content version. A newer edit or manual decision invalidates the event.
type ContentReviewRequestedEvent struct {
	Topic   bool
	ID      uint64
	Version uint64
}

func handleContentReviewRequested(_ context.Context, event *ContentReviewRequestedEvent) error {
	if event == nil || event.ID == 0 {
		return nil
	}
	if event.Topic {
		return reviewTopicContent(event)
	}
	return reviewPostContent(event)
}

func reviewTopicContent(event *ContentReviewRequestedEvent) error {
	topic := topics.Get(event.ID)
	if topic.Id == 0 || topic.ModerationVersion != event.Version || topic.ModerationStatus != "pending" {
		return nil
	}
	post := posts.Get(topic.FirstPostId)
	if post.Id == 0 {
		return nil
	}
	firstPublication := topic.PublishedAt == nil
	// Pending is assigned only after an explicit publication request.
	topic.Status = 1
	contentmoderationservice.ReviewTopic(&topic, &post)
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{
		Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds,
		ExpectedVersion: &event.Version, ReviewOnly: true,
	}); err != nil {
		return err
	}
	fileusageservice.ReplaceTopic(topic.Id, topic.UserId, post.Content)
	hotdataserve.ClearTopicCategoryCache()
	PublishTopicReviewResult(&topic, &post, firstPublication)
	return nil
}

func reviewPostContent(event *ContentReviewRequestedEvent) error {
	post := posts.Get(event.ID)
	if post.Id == 0 || post.ModerationVersion != event.Version || post.ModerationStatus != "pending" {
		return nil
	}
	topic := topics.Get(post.TopicId)
	if topic.Id == 0 {
		return nil
	}
	wasVisible, wasPublished := post.ProcessStatus == 0, post.WasPublished()
	contentmoderationservice.ReviewPost(&post)
	if err := posts.SaveReviewed(&post, event.Version); err != nil {
		return err
	}
	if wasVisible != (post.ProcessStatus == 0) {
		postservice.SyncTopicPostStats(topic, post, post.ProcessStatus != 0)
	}
	fileusageservice.ReplacePost(post.Id, post.UserId, post.Content)
	hotdataserve.ClearTopicListCache()
	if !wasPublished && post.ProcessStatus == 0 {
		PublishVisiblePost(topic, post)
	}
	return nil
}

// PublishTopicReviewResult dispatches the domain event produced by a topic
// write or review result.
func PublishTopicReviewResult(topic *topics.Entity, post *posts.Entity, firstPublication bool) {
	userservice.InvalidateUserPublicProfileCache(topic.UserId)
	if firstPublication && topic.Status == 1 && topic.ProcessStatus == 0 {
		userStatistics.WriteTopic(topic.UserId)
		eventbus.Publish(context.Background(), &TopicPublishedEvent{Topic: topic, FirstPost: post})
		return
	}
	eventbus.Publish(context.Background(), &TopicUpdatedEvent{Topic: topic, FirstPost: post})
}

// PublishVisiblePost dispatches the domain event produced when a reply first
// becomes publicly visible.
func PublishVisiblePost(topic topics.Entity, post posts.Entity) {
	if topic.Status != 1 || topic.ProcessStatus != 0 || post.ProcessStatus != 0 {
		return
	}
	userStatistics.WriteComment(post.UserId)
	userservice.InvalidateUserPublicProfileCache(post.UserId)
	hotdataserve.ClearTopicListCache()
	var parentUserID uint64
	if post.ReplyToPostId != 0 {
		parentUserID = posts.Get(post.ReplyToPostId).UserId
	}
	eventbus.Publish(context.Background(), &CommentCreatedEvent{
		TopicId: topic.Id, PostId: post.Id, PostNo: post.PostNo, UserId: post.UserId,
		Content: post.Content, TopicAuthorId: topic.UserId, ReplyToPostId: post.ReplyToPostId, ReplyToPostAuthorId: parentUserID,
	})
}
