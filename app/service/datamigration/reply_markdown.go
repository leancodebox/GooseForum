package datamigration

import (
	"log/slog"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
)

type ReplyMarkdownResult struct {
	Processed int
	Skipped   int
	Failed    int
}

func RebuildReplyMarkdown() ReplyMarkdownResult {
	conn := db.Connect()
	var startId uint64
	const limit = 200
	version := markdown2html.GetPostVersion()
	result := ReplyMarkdownResult{}
	if !conn.Migrator().HasTable("reply") {
		return result
	}

	for {
		var replyList []struct {
			Id              uint64
			Content         string
			RenderedHTML    string
			RenderedVersion uint32
		}
		if err := conn.Table("reply").
			Select("id", "content", "rendered_html", "rendered_version").
			Where("id > ?", startId).
			Order("id asc").
			Limit(limit).
			Find(&replyList).Error; err != nil {
			result.Failed++
			slog.Error("reply markdown scan failed", "err", err)
			return result
		}
		for _, item := range replyList {
			if startId < item.Id {
				startId = item.Id
			}
			if item.RenderedHTML != "" && item.RenderedVersion >= version {
				result.Skipped++
				continue
			}
			renderedHTML := markdown2html.PostMarkdownToHTML(item.Content)
			if err := conn.Table("reply").
				Where("id = ?", item.Id).
				Updates(map[string]any{
					"rendered_html":    renderedHTML,
					"rendered_version": version,
				}).Error; err != nil {
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
