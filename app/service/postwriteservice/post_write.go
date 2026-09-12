package postwriteservice

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/contentmoderationservice"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/topicunseenservice"
)

var (
	ErrPostNotFound      = errors.New("post not found")
	ErrTopicUnavailable  = errors.New("topic unavailable")
	ErrParentPostMissing = errors.New("parent post missing")
	ErrOwnerMismatch     = errors.New("post owner mismatch")
)

type CreateInput struct {
	UserID        uint64
	TopicID       uint64
	Content       string
	ReplyToPostID uint64
}

func Create(input CreateInput) (posts.Entity, error) {
	topic := topics.GetSimple(input.TopicID)
	if err := authorizePublishedTopic(input.UserID, topic, accesscontrol.CapabilityReply); err != nil {
		return posts.Entity{}, ErrTopicUnavailable
	}
	var parent posts.Entity
	if input.ReplyToPostID > 0 {
		parent = posts.Get(input.ReplyToPostID)
		if parent.Id == 0 || parent.TopicId != input.TopicID {
			return posts.Entity{}, ErrParentPostMissing
		}
	}

	analysis := markdown2html.AnalyzePostContent(input.Content)
	post := posts.Entity{
		TopicId: input.TopicID, Content: input.Content, RenderedHTML: analysis.RenderedHTML,
		RenderedVersion: markdown2html.GetPostVersion(), UserId: input.UserID, ReplyToPostId: input.ReplyToPostID,
	}
	if err := postservice.CreateTopicPost(&post, topic); err != nil {
		return posts.Entity{}, err
	}
	if err := topicunseenservice.MarkVisited(input.UserID, topic.Id, post.Id, time.Now()); err != nil {
		slog.Warn("mark created post visited failed", "userId", input.UserID, "topicId", topic.Id, "postId", post.Id, "error", err)
	}
	if len(analysis.ImageURLs) > 0 {
		fileusageservice.ReplacePostImages(post.Id, input.UserID, analysis.ImageURLs)
	}
	if post.ProcessStatus == 0 {
		eventhandlers.PublishVisiblePost(topic, post, parent.UserId)
	}
	enqueueReview(post.Id, post.ModerationVersion, post.ModerationStatus)
	return post, nil
}

type UpdateInput struct {
	UserID  uint64
	PostID  uint64
	Content string
}

func Update(input UpdateInput) (posts.Entity, error) {
	post := posts.Get(input.PostID)
	if post.Id == 0 || post.PostNo <= 1 {
		return posts.Entity{}, ErrPostNotFound
	}
	topic := topics.GetSimple(post.TopicId)
	if err := authorizePublishedTopic(input.UserID, topic, accesscontrol.CapabilityRead); err != nil {
		return posts.Entity{}, ErrTopicUnavailable
	}
	if post.UserId != input.UserID {
		return posts.Entity{}, ErrOwnerMismatch
	}

	expectedVersion := post.ModerationVersion
	wasPublished := post.WasPublished()
	wasVisible := post.ProcessStatus == 0
	analysis := markdown2html.AnalyzePostContent(input.Content)
	post.Content = input.Content
	post.RenderedHTML = analysis.RenderedHTML
	post.RenderedVersion = markdown2html.GetPostVersion()
	contentmoderationservice.PreparePost(&post)
	if err := posts.SaveReviewed(&post, expectedVersion); err != nil {
		return posts.Entity{}, err
	}
	fileusageservice.ReplacePostImages(post.Id, input.UserID, analysis.ImageURLs)
	if wasVisible != (post.ProcessStatus == 0) {
		postservice.SyncTopicPostStats(topic, post, post.ProcessStatus != 0)
		hotdataserve.ClearTopicListCache()
		if post.ProcessStatus == 0 && !wasPublished {
			eventhandlers.PublishVisiblePost(topic, post)
		}
	}
	enqueueReview(post.Id, post.ModerationVersion, post.ModerationStatus)
	return post, nil
}

func authorizePublishedTopic(userID uint64, topic topics.Entity, required accesscontrol.Capability) error {
	if topic.Id == 0 || topic.Status != 1 || topic.ProcessStatus != 0 {
		return ErrTopicUnavailable
	}
	snapshot, err := accesscontrol.Resolve(userID)
	if err != nil {
		return err
	}
	if required == accesscontrol.CapabilityReply {
		if snapshot.CanReplyCategory(topic.MainCategoryId) {
			return nil
		}
	} else if snapshot.CanReadCategory(topic.MainCategoryId) {
		return nil
	}
	return ErrTopicUnavailable
}

func enqueueReview(id, version uint64, status string) {
	if status == "pending" {
		eventbus.Publish(context.Background(), &eventhandlers.ContentReviewRequestedEvent{ID: id, Version: version})
	}
}
