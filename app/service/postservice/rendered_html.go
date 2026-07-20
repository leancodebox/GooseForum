package postservice

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
)

func EnsureRenderedHTML(entity *posts.Entity) string {
	html, err := ensureRenderedHTML(entity, posts.SaveNoUpdate)
	if err != nil {
		slog.Warn("save rebuilt post html failed", "postId", entity.Id, "error", err)
	}
	return html
}

func ensureRenderedHTML(entity *posts.Entity, save func(*posts.Entity) error) (string, error) {
	if entity == nil || entity.Id == 0 {
		return "", nil
	}
	if entity.RenderedVersion >= markdown2html.GetPostVersion() && entity.RenderedHTML != "" {
		return entity.RenderedHTML, nil
	}

	entity.RenderedHTML = markdown2html.PostMarkdownToHTML(entity.Content)
	entity.RenderedVersion = markdown2html.GetPostVersion()
	if err := save(entity); err != nil {
		return entity.RenderedHTML, err
	}
	return entity.RenderedHTML, nil
}
