package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/service/badgeservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

func handleBadgePost(ctx context.Context, event *TopicPublishedEvent) error {
	_, userID, _ := event.Subject()
	if userID == 0 {
		return nil
	}
	checkAndInvalidateUserBadges(userID, badgeservice.TriggerPost)
	return nil
}

func checkAndInvalidateUserBadges(userID uint64, trigger badgeservice.Trigger) {
	before := len(badgeservice.GetUserBadges(userID))
	badgeservice.CheckAndGrant(userID, trigger)
	if len(badgeservice.GetUserBadges(userID)) != before {
		userservice.InvalidateUserPublicProfileCache(userID)
	}
}

func handleBadgeComment(ctx context.Context, event *CommentCreatedEvent) error {
	checkAndInvalidateUserBadges(event.UserId, badgeservice.TriggerComment)
	return nil
}

func handleBadgeLike(ctx context.Context, event *TopicLikedEvent) error {
	checkAndInvalidateUserBadges(event.LikerId, badgeservice.TriggerLike)
	checkAndInvalidateUserBadges(event.UserId, badgeservice.TriggerLike)
	return nil
}

func handleBadgeFollow(ctx context.Context, event *UserFollowedEvent) error {
	checkAndInvalidateUserBadges(event.UserId, badgeservice.TriggerFollow)
	return nil
}
