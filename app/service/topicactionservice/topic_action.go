package topicactionservice

import (
	"context"
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

var ErrTopicUnavailable = errors.New("topic unavailable")

func SetLiked(userID, topicID uint64, liked bool) error {
	topic, state, err := loadActionState(userID, topicID)
	if err != nil || (state.Id == 0 && !liked) || (state.Id != 0 && (state.LikedAt != nil) == liked) {
		return err
	}
	if !topicUserAction.SetLiked(userID, topic.Id, liked) {
		return nil
	}
	if liked {
		topics.IncrementLike(topic.Id)
		userStatistics.LikeTopic(topic.UserId)
		userStatistics.GivenLike(userID)
		eventbus.Publish(context.Background(), &eventhandlers.TopicLikedEvent{
			UserId: topic.UserId, TopicId: topic.Id, Title: topic.Title, LikerId: userID,
		})
	} else {
		topics.DecrementLike(topic.Id)
		userStatistics.CancelLikeTopic(topic.UserId)
		userStatistics.CancelGivenLike(userID)
	}
	userservice.InvalidateUserPublicProfileCache(topic.UserId)
	userservice.InvalidateUserPublicProfileCache(userID)
	return nil
}

func SetBookmarked(userID, topicID uint64, bookmarked bool) error {
	_, state, err := loadActionState(userID, topicID)
	if err != nil || (state.Id == 0 && !bookmarked) || (state.Id != 0 && (state.BookmarkedAt != nil) == bookmarked) {
		return err
	}
	if !topicUserAction.SetBookmarked(userID, topicID, bookmarked) {
		return nil
	}
	if bookmarked {
		userStatistics.Collection(userID)
	} else {
		userStatistics.CancelCollection(userID)
	}
	userservice.InvalidateUserPublicProfileCache(userID)
	return nil
}

func SetWatched(userID, topicID uint64, watched bool) error {
	_, state, err := loadActionState(userID, topicID)
	if err != nil || (state.Id == 0 && !watched) || (state.Id != 0 && (state.WatchedAt != nil) == watched) {
		return err
	}
	topicUserAction.SetWatched(userID, topicID, watched)
	return nil
}

func loadActionState(userID, topicID uint64) (topics.InteractionTarget, topicUserAction.Entity, error) {
	topic := topics.GetInteractionTarget(topicID)
	if topic.Id == 0 || topic.Status != 1 || topic.ProcessStatus != 0 {
		return topics.InteractionTarget{}, topicUserAction.Entity{}, ErrTopicUnavailable
	}
	snapshot, err := accesscontrol.Resolve(userID)
	if err != nil || !snapshot.CanReadCategory(topic.MainCategoryId) {
		return topics.InteractionTarget{}, topicUserAction.Entity{}, ErrTopicUnavailable
	}
	return topic, topicUserAction.GetByTopicId(userID, topic.Id), nil
}
