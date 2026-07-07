package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
)

// ArticlePublishedEvent 文章发布事件
type ArticlePublishedEvent struct {
	Topic     *topics.Entity
	FirstPost *posts.Entity
}

func (event *ArticlePublishedEvent) Subject() (uint64, uint64, string) {
	if event == nil {
		return 0, 0, ""
	}
	if event.Topic != nil {
		return event.Topic.Id, event.Topic.UserId, event.Topic.Title
	}
	return 0, 0, ""
}

// handleArticlePublished 更新已发布文章搜索索引
func handleArticlePublished(ctx context.Context, event *ArticlePublishedEvent) error {
	if event == nil {
		return nil
	}
	if event.Topic != nil {
		_, err := searchservice.BuildSingleTopicSearchDocument(event.Topic, event.FirstPost)
		return err
	}
	return nil
}

// ArticleUpdatedEvent 文章更新事件
type ArticleUpdatedEvent struct {
	Topic     *topics.Entity
	FirstPost *posts.Entity
}

// handleArticleUpdated 更新文章搜索索引
func handleArticleUpdated(ctx context.Context, event *ArticleUpdatedEvent) error {
	if event == nil {
		return nil
	}
	if event.Topic != nil {
		_, err := searchservice.BuildSingleTopicSearchDocument(event.Topic, event.FirstPost)
		return err
	}
	return nil
}
