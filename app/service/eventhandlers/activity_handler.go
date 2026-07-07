package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
)

// handleActivitySignUp 记录注册行为
func handleActivitySignUp(ctx context.Context, event *UserSignUpEvent) error {
	return userActivities.Record(event.UserId, userActivities.ActionSignUp, userActivities.SubjectUser, event.UserId, "注册了账号")
}

// handleActivityPost 记录发帖行为
func handleActivityPost(ctx context.Context, event *TopicPublishedEvent) error {
	topicID, userID, title := event.Subject()
	if topicID == 0 || userID == 0 {
		return nil
	}
	return userActivities.Record(userID, userActivities.ActionPost, userActivities.SubjectTopic, topicID, title)
}

// handleActivityLike 记录点赞行为
func handleActivityLike(ctx context.Context, event *TopicLikedEvent) error {
	return userActivities.Record(event.LikerId, userActivities.ActionLike, userActivities.SubjectTopic, event.TopicId, event.Title)
}

// handleActivityFollow 记录关注行为
func handleActivityFollow(ctx context.Context, event *UserFollowedEvent) error {
	return userActivities.Record(event.FollowerId, userActivities.ActionFollow, userActivities.SubjectUser, event.UserId, "Followed a user")
}

// handleActivityReply 记录回复行为
func handleActivityReply(ctx context.Context, event *CommentCreatedEvent) error {
	return userActivities.Record(event.UserId, userActivities.ActionComment, userActivities.SubjectPost, event.PostId, TakeUpTo64Chars(event.Content))
}
