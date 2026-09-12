package topicwriteservice

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/contentmoderationservice"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"github.com/leancodebox/GooseForum/app/service/topicservice"
	"github.com/leancodebox/GooseForum/app/service/topicunseenservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

var (
	ErrTopicNotFound    = errors.New("topic not found")
	ErrOwnerMismatch    = errors.New("topic owner mismatch")
	ErrDailyLimit       = errors.New("daily topic limit reached")
	ErrPermissionDenied = errors.New("topic write permission denied")
)

type WriteInput struct {
	UserID      uint64
	TopicID     uint64
	Title       string
	Content     string
	CategoryIDs []uint64
	Status      int8
	DailyLimit  int
}

type WriteResult struct {
	ID               uint64
	ModerationStatus string
	TopicStatus      int8
}

type writeState struct {
	topic               topics.Entity
	firstPost           posts.Entity
	isNew               bool
	firstPublication    bool
	wasPublished        bool
	wasCounted          bool
	previousCategoryIDs []uint64
	expectedVersion     uint64
	imageURLs           []string
}

func Write(input WriteInput) (WriteResult, error) {
	state, err := loadWriteState(input)
	if err != nil {
		return WriteResult{}, err
	}
	categoryIDs, err := prepareWrite(state, input)
	if err != nil {
		return WriteResult{}, err
	}
	if err := persistWrite(state, categoryIDs); err != nil {
		return WriteResult{}, err
	}
	runWriteSideEffects(state, input.UserID)
	return WriteResult{
		ID: state.topic.Id, ModerationStatus: state.topic.ModerationStatus, TopicStatus: state.topic.Status,
	}, nil
}

func loadWriteState(input WriteInput) (*writeState, error) {
	state := &writeState{isNew: input.TopicID == 0, firstPublication: true}
	if !state.isNew {
		state.topic = topics.Get(input.TopicID)
		if state.topic.Id == 0 {
			return nil, ErrTopicNotFound
		}
		if state.topic.UserId != input.UserID {
			return nil, ErrOwnerMismatch
		}
		state.firstPublication = state.topic.PublishedAt == nil
		state.expectedVersion = state.topic.ModerationVersion
		state.wasPublished = state.topic.Status == 1 && state.topic.ProcessStatus == 0
		state.wasCounted = state.wasPublished
		state.previousCategoryIDs = append([]uint64(nil), state.topic.CategoryIds...)
		state.firstPost = posts.Get(state.topic.FirstPostId)
		if state.firstPost.Id == 0 {
			state.firstPost, _ = posts.GetByTopicPostNoAtOrAfter(state.topic.Id, 1)
		}
		if state.firstPost.Id == 0 {
			return nil, ErrTopicNotFound
		}
	} else {
		if input.DailyLimit > 0 && topics.CantWriteNew(input.UserID, int64(input.DailyLimit)) {
			return nil, ErrDailyLimit
		}
		state.topic.UserId = input.UserID
	}
	return state, nil
}

func prepareWrite(state *writeState, input WriteInput) ([]uint64, error) {
	categoryIDs, err := authorizeCategories(input.UserID, &state.topic, input.CategoryIDs, state.isNew, !state.wasPublished && input.Status == 1)
	if err != nil {
		return nil, errors.Join(ErrPermissionDenied, err)
	}
	analysis := markdown2html.AnalyzeContent(input.Content, 200)
	state.imageURLs = analysis.ImageURLs
	state.topic.CategoryIds = categoryIDs
	state.topic.Status = input.Status
	state.topic.Title = input.Title
	state.topic.Excerpt = analysis.Description
	state.topic.FirstImageURL = analysis.FirstImageURL

	if state.isNew {
		state.topic.Posters = []topics.Poster{{UserID: input.UserID}}
		state.firstPost = posts.Entity{
			UserId: input.UserID, Content: input.Content,
			RenderedVersion: markdown2html.GetPostVersion(),
		}
	} else {
		state.firstPost.Content = input.Content
		state.firstPost.RenderedHTML = ""
		state.firstPost.RenderedVersion = markdown2html.GetPostVersion()
	}
	contentmoderationservice.PrepareTopic(&state.topic, &state.firstPost)
	return categoryIDs, nil
}

func persistWrite(state *writeState, categoryIDs []uint64) error {
	write := topicservice.FirstPostWrite{
		Topic: &state.topic, FirstPost: &state.firstPost, CategoryIDs: categoryIDs, Create: state.isNew,
	}
	if !state.isNew {
		write.ExpectedVersion = &state.expectedVersion
	}
	return topicservice.SaveTopicAndFirstPost(write)
}

func runWriteSideEffects(state *writeState, userID uint64) {
	if !state.isNew || len(state.imageURLs) > 0 {
		fileusageservice.ReplaceTopicImages(state.topic.Id, userID, state.imageURLs)
	}
	if state.isNew {
		userservice.InvalidateUserPublicProfileCache(userID)
	}
	hotdataserve.ClearTopicWriteCaches(categoryCountsChanged(state.wasCounted, state.previousCategoryIDs, state.topic))
	eventhandlers.PublishTopicReviewResult(&state.topic, &state.firstPost, state.firstPublication)
	if state.isNew {
		if err := topicunseenservice.MarkVisited(userID, state.topic.Id, state.firstPost.Id, time.Now()); err != nil {
			slog.Warn("mark created topic visited failed", "userId", userID, "topicId", state.topic.Id, "error", err)
		}
	}
	enqueueReview(state.topic.Id, state.topic.ModerationVersion, state.topic.ModerationStatus)
}

func UpdateStatus(userID, topicID uint64, nextStatus int8) error {
	topic := topics.Get(topicID)
	if topic.Id == 0 {
		return ErrTopicNotFound
	}
	if topic.UserId != userID {
		return ErrOwnerMismatch
	}
	publishing := topic.Status != 1 && nextStatus == 1
	if _, err := authorizeCategories(userID, &topic, topic.CategoryIds, false, publishing); err != nil {
		return errors.Join(ErrPermissionDenied, err)
	}
	if topic.Status == nextStatus && !(nextStatus == 0 && topic.ModerationStatus == "pending") {
		return nil
	}
	firstPost := posts.Get(topic.FirstPostId)
	if firstPost.Id == 0 {
		return ErrTopicNotFound
	}
	firstPublication := topic.PublishedAt == nil
	wasCounted := topic.Status == 1 && topic.ProcessStatus == 0
	expectedVersion := topic.ModerationVersion
	topic.Status = nextStatus
	contentmoderationservice.PrepareTopic(&topic, &firstPost)
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{
		Topic: &topic, FirstPost: &firstPost, CategoryIDs: topic.CategoryIds, ExpectedVersion: &expectedVersion,
	}); err != nil {
		return err
	}
	hotdataserve.ClearTopicWriteCaches(wasCounted != (topic.Status == 1 && topic.ProcessStatus == 0))
	eventhandlers.PublishTopicReviewResult(&topic, &firstPost, firstPublication)
	enqueueReview(topic.Id, topic.ModerationVersion, topic.ModerationStatus)
	return nil
}

func authorizeCategories(userID uint64, topic *topics.Entity, next []uint64, newTopic, publishing bool) ([]uint64, error) {
	actor, err := accesscontrol.Resolve(userID)
	if err != nil {
		return nil, err
	}
	everyone, err := accesscontrol.Resolve(0)
	if err != nil {
		return nil, err
	}
	categoryIDs, err := accesscontrol.ValidateTopicCategoryWrite(actor, everyone, accesscontrol.TopicCategoryWrite{
		Current: topic.CategoryIds, Next: next, Publishing: publishing, NewTopic: newTopic,
	})
	if err != nil {
		return nil, err
	}
	if !category.AllExist(categoryIDs) {
		return nil, accesscontrol.ErrCategoryPermissionDenied
	}
	return categoryIDs, nil
}

func categoryCountsChanged(wasCounted bool, previousCategoryIDs []uint64, topic topics.Entity) bool {
	isCounted := topic.Status == 1 && topic.ProcessStatus == 0
	return wasCounted != isCounted || (isCounted && !slices.Equal(previousCategoryIDs, topic.CategoryIds))
}

func enqueueReview(id, version uint64, status string) {
	if status == "pending" {
		eventbus.Publish(context.Background(), &eventhandlers.ContentReviewRequestedEvent{Topic: true, ID: id, Version: version})
	}
}
