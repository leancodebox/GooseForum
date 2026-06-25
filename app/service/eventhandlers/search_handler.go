package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
)

// ArticlePublishedEvent 文章发布事件
type ArticlePublishedEvent struct {
	Article *articles.Entity
}

// handleArticlePublished 更新已发布文章搜索索引
func handleArticlePublished(ctx context.Context, event *ArticlePublishedEvent) error {
	_, err := searchservice.BuildSingleArticleSearchDocument(event.Article)
	return err
}

// ArticleUpdatedEvent 文章更新事件
type ArticleUpdatedEvent struct {
	Article *articles.Entity
}

// handleArticleUpdated 更新文章搜索索引
func handleArticleUpdated(ctx context.Context, event *ArticleUpdatedEvent) error {
	_, err := searchservice.BuildSingleArticleSearchDocument(event.Article)
	return err
}
