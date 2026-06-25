package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/service/pointservice"
)

// handlePointArticlePublished 发帖获得积分
func handlePointArticlePublished(ctx context.Context, event *ArticlePublishedEvent) error {
	pointservice.RewardPoints(event.Article.UserId, 10, pointservice.RewardPoints4WriteArticles)
	return nil
}

// handlePointCommentCreated 评论获得积分
func handlePointCommentCreated(ctx context.Context, event *CommentCreatedEvent) error {
	pointservice.RewardPoints(event.UserId, 2, pointservice.RewardPoints4Reply)
	return nil
}
