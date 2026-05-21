package datamigration

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
)

type ReplyMarkdownResult struct {
	Processed int
	Skipped   int
	Failed    int
}

func RebuildReplyMarkdown() ReplyMarkdownResult {
	var startId uint64
	const limit = 200
	version := markdown2html.GetCommentVersion()
	result := ReplyMarkdownResult{}

	for {
		replyList := reply.QueryById(startId, limit)
		for _, item := range replyList {
			if item == nil {
				continue
			}
			if startId < item.Id {
				startId = item.Id
			}
			if item.RenderedHTML != "" && item.RenderedVersion >= version {
				result.Skipped++
				continue
			}
			item.RenderedHTML = markdown2html.CommentMarkdownToHTML(item.Content)
			item.RenderedVersion = version
			if err := reply.SaveNoUpdate(item); err != nil {
				result.Failed++
				slog.Error("reply markdown rebuild failed", "replyId", item.Id, "err", err)
				continue
			}
			result.Processed++
		}

		if len(replyList) < limit {
			break
		}
	}

	return result
}
