package eventhandlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
)

// ArticlePublishedEvent 文章发布事件
type ArticlePublishedEvent struct {
	Article *articles.Entity
}

// ArticlePublishedHandler 文章发布处理器
func NewArticlePublishedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ArticlePublishedHandler",
		func(ctx context.Context, event *ArticlePublishedEvent) error {
			_, err := searchservice.BuildSingleArticleSearchDocument(event.Article)
			return err
		},
	)
}

// ArticleUpdatedEvent 文章更新事件
type ArticleUpdatedEvent struct {
	Article *articles.Entity
}

// ArticleUpdatedHandler 文章更新处理器
func NewArticleUpdatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ArticleUpdatedHandler",
		func(ctx context.Context, event *ArticleUpdatedEvent) error {
			_, err := searchservice.BuildSingleArticleSearchDocument(event.Article)
			return err
		},
	)
}
