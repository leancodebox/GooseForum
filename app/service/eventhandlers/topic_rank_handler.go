package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
)

func handleTopicRankRequested(ctx context.Context, event *topicrank.TopicRankRequested) error {
	// Returning errors lets the event bus apply its existing retry policy.
	return topicrank.MarkAt(ctx, event.TopicID, event.OccurredAt)
}
