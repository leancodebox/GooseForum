package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/service/pointservice"
)

// handlePointTopicPublished 发帖获得积分
func handlePointTopicPublished(ctx context.Context, event *TopicPublishedEvent) error {
	_, userID, _ := event.Subject()
	if userID == 0 {
		return nil
	}
	pointservice.RewardPoints(userID, 10, pointservice.PointsActionTopicPublished)
	return nil
}

// handlePointCommentCreated 评论获得积分
func handlePointCommentCreated(ctx context.Context, event *CommentCreatedEvent) error {
	pointservice.RewardPoints(event.UserId, 2, pointservice.PointsActionPostCreated)
	return nil
}
