package eventhandlers

import (
	"context"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
)

// handleStatsSignUp 记录注册统计
func handleStatsSignUp(ctx context.Context, event *UserSignUpEvent) error {
	return dailyStats.Increment(time.Now(), dailyStats.StatTypeRegCount, 1)
}

// handleStatsPost 记录发帖统计
func handleStatsPost(ctx context.Context, event *TopicPublishedEvent) error {
	return dailyStats.Increment(time.Now(), dailyStats.StatTypeArticleCount, 1)
}

// handleStatsReply 记录回复统计
func handleStatsReply(ctx context.Context, event *CommentCreatedEvent) error {
	return dailyStats.Increment(time.Now(), dailyStats.StatTypeReplyCount, 1)
}
