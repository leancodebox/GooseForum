package userfollowservice

import (
	"context"
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

var ErrUserNotFound = errors.New("user not found")

func SetFollowed(userID, followedUserID uint64, followed bool) error {
	target, _ := users.GetIdentity(followedUserID)
	if target.Id == 0 {
		return ErrUserNotFound
	}
	relation := userFollow.GetByUserId(userID, followedUserID)
	if relation.Id == 0 {
		relation.UserId = userID
		relation.FollowUserId = followedUserID
	}
	targetStatus := 0
	if followed {
		targetStatus = 1
	}
	if relation.Status == targetStatus {
		return nil
	}
	relation.Status = targetStatus
	if userFollow.SaveOrCreateById(&relation) == 0 {
		return nil
	}
	if followed {
		userStatistics.Following(userID)
		userStatistics.Follower(followedUserID)
		follower, _ := users.GetIdentity(userID)
		eventbus.Publish(context.Background(), &eventhandlers.UserFollowedEvent{
			UserId: followedUserID, FollowerId: userID, FollowerName: follower.Username,
		})
	} else {
		userStatistics.CancelFollowing(userID)
		userStatistics.CancelFollower(followedUserID)
	}
	userservice.InvalidateUserPublicProfileCache(userID)
	userservice.InvalidateUserPublicProfileCache(followedUserID)
	return nil
}
